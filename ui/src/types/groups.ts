export interface Group {
  id: string;
  name: string;
  labels: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface CreateGroupRequest {
  name: string;
  labels?: Record<string, string>;
}
