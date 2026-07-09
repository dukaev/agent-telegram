export type Safeties = {
  read: boolean;
  write: boolean;
  destructive: boolean;
  paid: boolean;
};

export type PeerTypes = {
  users: boolean;
  groups: boolean;
  channels: boolean;
  bots: boolean;
};

export type Policy = {
  version: number;
  safeties: Safeties;
  peerTypes: PeerTypes;
  allowPeers?: string[];
  denyPeers?: string[];
};

export type AuthMode = "qr" | "phone" | "code" | "password" | "setup" | "done";

export type AuthAPI = {
  appId: number;
  default: boolean;
  canEdit: boolean;
};

export type AuthMock = {
  enabled: boolean;
  code?: string;
  password?: string;
};

export type AuthState = {
  title: string;
  message?: string;
  error?: string;
  mode: AuthMode;
  completed: boolean;
  phone?: string;
  hint?: string;
  qrImage?: string;
  qrLink?: string;
  expires?: string;
  refresh?: number;
  api: AuthAPI;
  policy: Policy;
  mock?: AuthMock;
};

export type PeerOption = {
  peer: string;
  title: string;
  username?: string;
  type: "user" | "group" | "channel" | "bot" | "";
  id?: number;
};

export type PeersState = {
  peers: PeerOption[];
  count: number;
  loaded: boolean;
  loading: boolean;
  error?: string;
};

export const defaultPolicy: Policy = {
  version: 1,
  safeties: {
    read: true,
    write: true,
    destructive: false,
    paid: false,
  },
  peerTypes: {
    users: true,
    groups: true,
    channels: true,
    bots: true,
  },
  allowPeers: [],
  denyPeers: [],
};

export const peerTypeLabels: Record<string, string> = {
  all: "All",
  user: "People",
  group: "Groups",
  channel: "Channels",
  bot: "Bots",
};
