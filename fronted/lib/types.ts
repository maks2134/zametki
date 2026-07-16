export type Category =
  | "idea"
  | "date"
  | "gift"
  | "movie"
  | "travel"
  | "thought"
  | "other";

export interface Member {
  id: string;
  name: string;
  color: string;
  joinedAt: string;
}

export interface Room {
  id: string;
  code: string;
  title: string;
  createdAt: string;
  members: Member[];
}

export interface Reaction {
  memberId: string;
  emoji: string;
}

export interface Note {
  id: string;
  roomId: string;
  authorId: string;
  title?: string;
  content: string;
  category: Category;
  color?: string;
  pinned: boolean;
  reactions: Reaction[];
  createdAt: string;
  updatedAt: string;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
  };
}

export type WSEvent =
  | { type: "note.created"; data: Note }
  | { type: "note.updated"; data: Note }
  | { type: "note.deleted"; data: { id: string } }
  | { type: "reaction.updated"; data: Note }
  | { type: "member.joined"; data: Member };

export const CATEGORIES: { value: Category | "all"; label: string }[] = [
  { value: "all", label: "Все" },
  { value: "idea", label: "Идеи" },
  { value: "date", label: "Свидания" },
  { value: "gift", label: "Подарки" },
  { value: "movie", label: "Кино" },
  { value: "travel", label: "Путешествия" },
  { value: "thought", label: "Мысли" },
  { value: "other", label: "Другое" },
];

export const MEMBER_COLORS = [
  "#e85d75",
  "#f4a261",
  "#2a9d8f",
  "#7b68ee",
  "#e76f51",
  "#48cae4",
];
