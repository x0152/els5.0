export interface paths {
    "/api/v1/onboarding/ack": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Mark unlocked achievements as seen */
        post: operations["onboardingAck"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/onboarding/progress": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Getting-started checklist and achievements with accumulated progress */
        get: operations["onboardingProgress"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/onboarding/tours": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Completed onboarding tour ids */
        get: operations["onboardingTours"];
        put?: never;
        /** Mark an onboarding tour as completed */
        post: operations["onboardingMarkTour"];
        /** Reset onboarding tours so they show again */
        delete: operations["onboardingResetTours"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        AckItemsInputBody: {
            /** @description Achievement item ids to mark as seen */
            ids: string[] | null;
        };
        AckItemsOutput: Record<string, never>;
        CheckResult: {
            error?: string;
            /** @example postgres */
            name: string;
            ok: boolean;
        };
        ErrorBody: {
            code: string;
            details?: components["schemas"]["ErrorDetail"][] | null;
            message: string;
        };
        ErrorDetail: {
            field?: string;
            message: string;
        };
        ErrorResponse: {
            error: components["schemas"]["ErrorBody"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        HealthOutput: {
            /** @example auth */
            module?: string;
            /** @example ok */
            status: string;
            /** Format: date-time */
            time: string;
        };
        MarkTourInputBody: {
            /** @description Tour id to mark as completed */
            id: string;
        };
        MarkTourOutput: Record<string, never>;
        Meta: {
            pagination?: components["schemas"]["Pagination"];
            request_id?: string;
        };
        OnboardingItemOutput: {
            acked: boolean;
            done: boolean;
            id: string;
            /** @enum {string} */
            kind: "checklist" | "achievement";
            metric: string;
            /** Format: int64 */
            threshold: number;
            /** Format: int64 */
            value: number;
        };
        OnboardingProgressOutput: {
            items: components["schemas"]["OnboardingItemOutput"][] | null;
        };
        Pagination: {
            has_more: boolean;
            /** Format: int64 */
            limit: number;
            /** Format: int64 */
            offset: number;
            /** Format: int64 */
            total: number;
        };
        ReadyOutput: {
            checks?: components["schemas"]["CheckResult"][] | null;
            /** @example auth */
            module?: string;
            ready: boolean;
            /** Format: date-time */
            time: string;
        };
        ResetToursOutput: Record<string, never>;
        SuccessBodyAckItemsOutput: {
            data: components["schemas"]["AckItemsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyMarkTourOutput: {
            data: components["schemas"]["MarkTourOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyOnboardingProgressOutput: {
            data: components["schemas"]["OnboardingProgressOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyResetToursOutput: {
            data: components["schemas"]["ResetToursOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyToursOutput: {
            data: components["schemas"]["ToursOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        ToursOutput: {
            ids: string[] | null;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaAckItemsInputBody = components['schemas']['AckItemsInputBody'];
export type SchemaAckItemsOutput = components['schemas']['AckItemsOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaMarkTourInputBody = components['schemas']['MarkTourInputBody'];
export type SchemaMarkTourOutput = components['schemas']['MarkTourOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaOnboardingItemOutput = components['schemas']['OnboardingItemOutput'];
export type SchemaOnboardingProgressOutput = components['schemas']['OnboardingProgressOutput'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaResetToursOutput = components['schemas']['ResetToursOutput'];
export type SchemaSuccessBodyAckItemsOutput = components['schemas']['SuccessBodyAckItemsOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyMarkTourOutput = components['schemas']['SuccessBodyMarkTourOutput'];
export type SchemaSuccessBodyOnboardingProgressOutput = components['schemas']['SuccessBodyOnboardingProgressOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodyResetToursOutput = components['schemas']['SuccessBodyResetToursOutput'];
export type SchemaSuccessBodyToursOutput = components['schemas']['SuccessBodyToursOutput'];
export type SchemaToursOutput = components['schemas']['ToursOutput'];
export type $defs = Record<string, never>;
export interface operations {
    onboardingAck: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["AckItemsInputBody"];
            };
        };
        responses: {
            /** @description No Content */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyAckItemsOutput"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    onboardingProgress: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyOnboardingProgressOutput"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    onboardingTours: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyToursOutput"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    onboardingMarkTour: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["MarkTourInputBody"];
            };
        };
        responses: {
            /** @description No Content */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyMarkTourOutput"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    onboardingResetTours: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description No Content */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyResetToursOutput"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
}
