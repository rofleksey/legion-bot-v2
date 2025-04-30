export interface User {
  login: string;
  displayName: string;
  profileImageUrl: string;
}

export interface Settings {
  disabled: boolean;
  language: string;
  killers: KillersSettings;
  chat: ChatSettings;
}

export interface ChatSettings {
  startKillerOnRaid: boolean;
  followRaids: boolean;
  followRaidsMessage: string;
}

export interface KillersSettings {
  general: GeneralKillerSettings;
  legion: LegionSettings;
  ghostface: GhostFaceSettings;
  doctor: DoctorSettings;
  pinhead: PinheadSettings;
}

export interface GeneralKillerSettings {
  delayBetweenKillers: number;
  delayAtTheStreamStart: number;
  minNumberOfViewers: number;
}

export interface LegionSettings {
  enabled: boolean;
  weight: number;
  bodyBlockSuccessChance: number;
  deepWoundTimeout: number;
  fatalHit: number;
  frenzyTimeout: number;
  hitChance: number;
  hookBanTime: number;
  lockerGrabChance: number;
  lockerStunChance: number;
  minDelayBetweenHits: number;
  palletStunChance: number;
  reactChance: number;
  bleedOutBanTime: number;
}

export interface GhostFaceSettings {
  enabled: boolean;
  weight: number;
  hookBanTime: number;
  reactChance: number;
  minDelayBetweenHits: number;
  timeout: number;
}

export interface DoctorSettings {
  enabled: boolean;
  weight: number;
  reactChance: number;
  minDelayBetweenHits: number;
  timeout: number;
}

export interface PinheadSettings {
  enabled: boolean;
  weight: number;
  showTopic: boolean;
  victimCount: number;
  deepWoundTimeout: number;
  bleedOutBanTime: number;
  topics: string;
  timeout: number;
}
