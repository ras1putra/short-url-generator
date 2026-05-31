// Wallet error codes
export const CHAIN_NOT_ADDED = 4902;
export const USER_REJECTED = 4001;

// Roles
export const ROLE_USER = "user";
export const ROLE_ADVERTISER = "advertiser";
export const ROLE_ADMIN = "admin";

export const ROLES = [ROLE_USER, ROLE_ADVERTISER, ROLE_ADMIN] as const;

// Expiry Units
export const EXPIRY_UNIT_MINUTES = "minutes";
export const EXPIRY_UNIT_HOURS = "hours";
export const EXPIRY_UNIT_DAYS = "days";

export const EXPIRY_UNITS = [
  EXPIRY_UNIT_MINUTES,
  EXPIRY_UNIT_HOURS,
  EXPIRY_UNIT_DAYS,
] as const;

// Pagination
export const DEFAULT_PAGE_SIZE = 5;
export const PAGE_SIZE_OPTIONS = [5, 10, 25, 50];

// API Endpoints
export const API_AUTH_LOGIN = "/api/auth/login";
export const API_AUTH_REGISTER = "/api/auth/register";
export const API_AUTH_LOGOUT = "/api/auth/logout";
export const API_AUTH_REFRESH = "/api/auth/refresh";
export const API_AUTH_UPGRADE = "/api/auth/upgrade";
export const API_AUTH_DOWNGRADE = "/api/auth/downgrade";
export const API_AUTH_GOOGLE_LOGIN = "/api/auth/google/login";
export const API_AUTH_ME = "/api/auth/me";
export const API_AUTH_SEND_VERIFICATION = "/api/auth/send-verification";
export const API_AUTH_VERIFY_EMAIL = "/api/auth/verify-email";
export const API_AUTH_FORGOT_PASSWORD = "/api/auth/forgot-password";
export const API_AUTH_RESET_PASSWORD = "/api/auth/reset-password";
export const API_LINKS = "/api/links";
export const API_ADS = "/api/ads";
export const API_MEDIA_UPLOAD = "/api/media/upload";
export const API_MEDIA_CROP_VIDEO = "/api/media/crop-video";
export const API_FAUCET_CLAIM = "/api/faucet/claim";
export const API_FAUCET_CONFIRM = "/api/faucet/confirm";
export const API_FAUCET_DEV_ETH = "/api/faucet/dev-eth";
export const API_FAUCET_HISTORY = "/api/faucet/history";
export const API_CONFIG = "/api/config";
export const API_CATEGORIES = "/api/categories";
export const API_WALLET = "/api/wallet";

// Routes
export const ROUTE_LOGIN = "/login";
export const ROUTE_REGISTER = "/register";
export const ROUTE_REGISTER_ADVERTISER = "/register/advertiser";
export const ROUTE_FORGOT_PASSWORD = "/forgot-password";
export const ROUTE_RESET_PASSWORD = "/reset-password";
export const ROUTE_VERIFY_EMAIL = "/verify-email";
export const ROUTE_DASHBOARD = "/dashboard";
export const ROUTE_CAMPAIGNS = "/dashboard/campaigns";
export const ROUTE_LINKS = "/dashboard/links";
export const ROUTE_WALLET = "/dashboard/wallet";
export const ROUTE_FAUCET = "/dashboard/faucet";
export const ROUTE_ADMIN_DASHBOARD = "/dashboard/admin";
export const ROUTE_SETTINGS = "/dashboard/settings";

// Faucet
export const FAUCET_AMOUNT = 20;

export const RATIO_TOLERANCE = 0.05;

// Media upload limits
export const MAX_IMAGE_SIZE = 5 * 1024 * 1024;
export const MAX_VIDEO_SIZE = 20 * 1024 * 1024;
export const MEDIA_ACCEPT_TYPES = {
  images: "image/png,image/jpeg,image/gif,image/webp",
  videos: "video/mp4,video/webm",
};

// Content type constants
export const CONTENT_TYPE_PNG = "image/png";
export const CONTENT_TYPE_JPEG = "image/jpeg";
export const CONTENT_TYPE_GIF = "image/gif";
export const CONTENT_TYPE_WEBP = "image/webp";
export const CONTENT_TYPE_MP4 = "video/mp4";
export const CONTENT_TYPE_WEBM = "video/webm";

// Confirmations
export const BLOCK_CONFIRMATIONS = 12;

// Ad Statuses
export const AD_STATUS_ACTIVE = "active";
export const AD_STATUS_PAUSED = "paused";
export const AD_STATUS_DELETED = "deleted";

// Transaction Statuses
export const TX_STATUS_PENDING = "PENDING";
export const TX_STATUS_CONFIRMED = "CONFIRMED";
export const TX_STATUS_FAILED = "FAILED";

// Transaction Types
export const TX_TYPE_EARNING = "EARNING";
export const TX_TYPE_AD_EARNING = "AD_EARNING";
export const TX_TYPE_AD_SPEND = "AD_SPEND";
export const TX_TYPE_DEPOSIT = "DEPOSIT";
export const TX_TYPE_WITHDRAWAL = "WITHDRAWAL";
export const TX_TYPE_WITHDRAWAL_FEE = "WITHDRAWAL_FEE";
export const TX_TYPE_FAUCET = "FAUCET";

// Ad Rejection Reasons
export const REJECT_REASON_HONEYPOT_HIT = "HONEYPOT_HIT";
export const REJECT_REASON_TOO_FAST = "TOO_FAST";
export const REJECT_REASON_NO_MOUSE_MOVEMENT = "NO_MOUSE_MOVEMENT";
export const REJECT_REASON_DUPLICATE_IP = "DUPLICATE_IP";
export const REJECT_REASON_DUPLICATE_FINGERPRINT = "DUPLICATE_FINGERPRINT";

// Ad Event Types
export const AD_EVENT_IMPRESSION = "IMPRESSION";
export const AD_EVENT_CLICK = "CLICK";
export const AD_EVENT_COMPLETION = "COMPLETION";
export const AD_EVENT_SKIP = "SKIP";
