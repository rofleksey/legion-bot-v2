export interface User {
  id: string
  login: string;
  displayName: string;
  profileImageUrl: string;
  email: string;
}

export interface Settings {
  disabled: boolean;
  language: string;
  killers: KillersSettings;
}

export interface KillersSettings {
  general: GeneralKillerSettings;
  legion: LegionSettings;
}

export interface GeneralKillerSettings {
  delayBetweenKillers: number;
}

export interface LegionSettings {
  enabled: boolean;
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
