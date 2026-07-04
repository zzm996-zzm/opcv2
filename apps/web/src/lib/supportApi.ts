import { apiRequest } from "./apiRequest";

export type SupportTicketInput = {
  topic: string;
  title: string;
  body: string;
};

export type SupportTicket = {
  id: number;
  user_id?: number;
  topic?: string;
  title: string;
  body?: string;
  status: string;
  created_at: string;
  updated_at?: string;
};

export const supportApi = {
  createTicket(input: SupportTicketInput) {
    return apiRequest<SupportTicket>("/api/v1/support/tickets", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  listTickets(limit = 20) {
    return apiRequest<{ tickets: SupportTicket[] }>(`/api/v1/support/tickets?limit=${limit}`);
  }
};
