export const ANALYTICS_EVENTS = {
  APP_INIT: "app_init",
  LOGIN_SUCCESS: "login_success",
  LOGIN_ERROR: "login_error",
  REGISTER_SUCCESS: "register_success",
  ROUTINE_COMPLETED: "routine_completed",
} as const;

export type AnalyticsEvent = typeof ANALYTICS_EVENTS[keyof typeof ANALYTICS_EVENTS];
