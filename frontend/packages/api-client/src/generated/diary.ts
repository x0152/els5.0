export interface paths {
    "/api/v1/diary/entries": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List past diary entries */
        get: operations["diaryListEntries"];
        put?: never;
        /** Submit today's entry and get the friend reply with corrections */
        post: operations["diarySubmitEntry"];
        /** Delete all diary entries of the account */
        delete: operations["diaryResetHistory"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/diary/entries/check": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Check a diary draft for grammar errors before submitting */
        post: operations["diaryCheckEntry"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/diary/today": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Today's diary state: question, warmup and streak */
        get: operations["diaryToday"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        CheckEntryInputBody: {
            /** @description Diary draft to check for grammar errors */
            text: string;
        };
        CheckEntryOutput: {
            errors: components["schemas"]["GrammarErrorOutput"][] | null;
            ok: boolean;
        };
        CheckResult: {
            error?: string;
            /** @example postgres */
            name: string;
            ok: boolean;
        };
        CorrectionOutput: {
            correction: string;
            description: string;
            fragment: string;
            sentence: string;
        };
        EntriesOutput: {
            items: components["schemas"]["EntryOutput"][] | null;
            /** Format: int32 */
            limit: number;
            /** Format: int32 */
            offset: number;
            /** Format: int64 */
            total: number;
        };
        EntryOutput: {
            corrections: components["schemas"]["CorrectionOutput"][] | null;
            /** Format: date-time */
            created_at: string;
            date: string;
            draft?: string;
            id: string;
            native_sample?: string;
            next_question?: string;
            question?: string;
            reply: string;
            /**
             * @description pending while the friend reply is being generated
             * @enum {string}
             */
            status: "pending" | "done";
            text: string;
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
        GrammarErrorOutput: {
            correction: string;
            explanation: string;
            original: string;
            type: string;
        };
        HealthOutput: {
            /** @example auth */
            module?: string;
            /** @example ok */
            status: string;
            /** Format: date-time */
            time: string;
        };
        Meta: {
            pagination?: components["schemas"]["Pagination"];
            request_id?: string;
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
        ResetHistoryOutput: Record<string, never>;
        SubmitEntryInputBody: {
            /** @description The first version of the text, before error fixes */
            draft?: string;
            /** @description The question the entry answers */
            question?: string;
            /** @description Diary entry text in English */
            text: string;
        };
        SuccessBodyCheckEntryOutput: {
            data: components["schemas"]["CheckEntryOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyEntriesOutput: {
            data: components["schemas"]["EntriesOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyEntryOutput: {
            data: components["schemas"]["EntryOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyResetHistoryOutput: {
            data: components["schemas"]["ResetHistoryOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyTodayOutput: {
            data: components["schemas"]["TodayOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        TodayOutput: {
            entry?: components["schemas"]["EntryOutput"];
            question: string;
            /** Format: int64 */
            streak: number;
            warmup: components["schemas"]["CorrectionOutput"][] | null;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaCheckEntryInputBody = components['schemas']['CheckEntryInputBody'];
export type SchemaCheckEntryOutput = components['schemas']['CheckEntryOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaCorrectionOutput = components['schemas']['CorrectionOutput'];
export type SchemaEntriesOutput = components['schemas']['EntriesOutput'];
export type SchemaEntryOutput = components['schemas']['EntryOutput'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaGrammarErrorOutput = components['schemas']['GrammarErrorOutput'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaResetHistoryOutput = components['schemas']['ResetHistoryOutput'];
export type SchemaSubmitEntryInputBody = components['schemas']['SubmitEntryInputBody'];
export type SchemaSuccessBodyCheckEntryOutput = components['schemas']['SuccessBodyCheckEntryOutput'];
export type SchemaSuccessBodyEntriesOutput = components['schemas']['SuccessBodyEntriesOutput'];
export type SchemaSuccessBodyEntryOutput = components['schemas']['SuccessBodyEntryOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodyResetHistoryOutput = components['schemas']['SuccessBodyResetHistoryOutput'];
export type SchemaSuccessBodyTodayOutput = components['schemas']['SuccessBodyTodayOutput'];
export type SchemaTodayOutput = components['schemas']['TodayOutput'];
export type $defs = Record<string, never>;
export interface operations {
    diaryListEntries: {
        parameters: {
            query?: {
                limit?: number;
                offset?: number;
            };
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
                    "application/json": components["schemas"]["SuccessBodyEntriesOutput"];
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
    diarySubmitEntry: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SubmitEntryInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyEntryOutput"];
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
    diaryResetHistory: {
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
                    "application/json": components["schemas"]["SuccessBodyResetHistoryOutput"];
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
    diaryCheckEntry: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CheckEntryInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyCheckEntryOutput"];
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
    diaryToday: {
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
                    "application/json": components["schemas"]["SuccessBodyTodayOutput"];
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
