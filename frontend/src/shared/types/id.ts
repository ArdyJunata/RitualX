declare const brand: unique symbol;
export type Brand<T, B extends string> = T & { readonly [brand]: B };

export type UserId = Brand<string, "UserId">;
export type RoutineId = Brand<string, "RoutineId">;

export const toUserId = (s: string): UserId => s as UserId;
export const toRoutineId = (s: string): RoutineId => s as RoutineId;
