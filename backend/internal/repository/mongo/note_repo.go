package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"zametka/internal/domain"
)

const notesCollection = "notes"

type NoteRepository struct {
	coll *mongo.Collection
}

func NewNoteRepository(db *mongo.Database) *NoteRepository {
	return &NoteRepository{coll: db.Collection(notesCollection)}
}

func (r *NoteRepository) Create(ctx context.Context, n *domain.Note) error {
	if n.Reactions == nil {
		n.Reactions = []domain.Reaction{}
	}
	_, err := r.coll.InsertOne(ctx, n)
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}
	return nil
}

func (r *NoteRepository) GetByID(ctx context.Context, roomID, id string) (*domain.Note, error) {
	var note domain.Note
	err := r.coll.FindOne(ctx, bson.M{"_id": id, "roomId": roomID}).Decode(&note)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, fmt.Errorf("find note: %w", err)
	}
	if note.Reactions == nil {
		note.Reactions = []domain.Reaction{}
	}
	return &note, nil
}

func (r *NoteRepository) List(ctx context.Context, f domain.NoteFilter) ([]domain.Note, error) {
	filter := bson.M{"roomId": f.RoomID}
	if f.Category != nil {
		filter["category"] = *f.Category
	}
	if f.Before != nil {
		filter["createdAt"] = bson.M{"$lt": *f.Before}
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit)

	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer cur.Close(ctx)

	notes := make([]domain.Note, 0, limit)
	for cur.Next(ctx) {
		var note domain.Note
		if err := cur.Decode(&note); err != nil {
			return nil, fmt.Errorf("decode note: %w", err)
		}
		if note.Reactions == nil {
			note.Reactions = []domain.Reaction{}
		}
		notes = append(notes, note)
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}
	return notes, nil
}

func (r *NoteRepository) Update(ctx context.Context, roomID, id string, upd domain.NoteUpdate) (*domain.Note, error) {
	set := bson.M{"updatedAt": time.Now().UTC()}
	if upd.Title != nil {
		set["title"] = *upd.Title
	}
	if upd.Content != nil {
		set["content"] = *upd.Content
	}
	if upd.Category != nil {
		set["category"] = *upd.Category
	}
	if upd.Color != nil {
		set["color"] = *upd.Color
	}
	if upd.Pinned != nil {
		set["pinned"] = *upd.Pinned
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var note domain.Note
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "roomId": roomID},
		bson.M{"$set": set},
		opts,
	).Decode(&note)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, fmt.Errorf("update note: %w", err)
	}
	if note.Reactions == nil {
		note.Reactions = []domain.Reaction{}
	}
	return &note, nil
}

func (r *NoteRepository) Delete(ctx context.Context, roomID, id string) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id, "roomId": roomID})
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrNoteNotFound
	}
	return nil
}

func (r *NoteRepository) UpsertReaction(ctx context.Context, roomID, noteID, memberID, emoji string) (*domain.Note, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"_id":                noteID,
		"roomId":             roomID,
		"reactions.memberId": memberID,
	}
	update := bson.M{
		"$set": bson.M{
			"reactions.$.emoji": emoji,
			"updatedAt":         now,
		},
	}

	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("upsert reaction update: %w", err)
	}

	if res.MatchedCount == 0 {
		pushFilter := bson.M{"_id": noteID, "roomId": roomID}
		pushUpdate := bson.M{
			"$push": bson.M{
				"reactions": domain.Reaction{MemberID: memberID, Emoji: emoji},
			},
			"$set": bson.M{"updatedAt": now},
		}
		pushRes, err := r.coll.UpdateOne(ctx, pushFilter, pushUpdate)
		if err != nil {
			return nil, fmt.Errorf("upsert reaction push: %w", err)
		}
		if pushRes.MatchedCount == 0 {
			return nil, domain.ErrNoteNotFound
		}
	}

	return r.GetByID(ctx, roomID, noteID)
}

func (r *NoteRepository) RemoveReaction(ctx context.Context, roomID, noteID, memberID string) (*domain.Note, error) {
	now := time.Now().UTC()
	filter := bson.M{"_id": noteID, "roomId": roomID}
	update := bson.M{
		"$pull": bson.M{
			"reactions": bson.M{"memberId": memberID},
		},
		"$set": bson.M{"updatedAt": now},
	}

	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("remove reaction: %w", err)
	}
	if res.MatchedCount == 0 {
		return nil, domain.ErrNoteNotFound
	}

	return r.GetByID(ctx, roomID, noteID)
}
