export interface paths {
    "/api/v1/practice/check": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Validate a free-form answer with the LLM */
        post: operations["checkPracticeAnswer"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/practice/variants/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Delete a generated practice variant */
        delete: operations["deletePracticeVariant"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/practice/{kind}/{number}/progress": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get saved answers and completion for a chapter variant */
        get: operations["getPracticeProgress"];
        /** Save answers and completion for a chapter variant */
        put: operations["savePracticeProgress"];
        post?: never;
        /** Reset saved answers for a chapter variant */
        delete: operations["resetPracticeProgress"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/practice/{kind}/{number}/variants": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List generated practice variants for a chapter */
        get: operations["listPracticeVariants"];
        put?: never;
        /** Generate a new practice variant with the LLM */
        post: operations["generatePracticeVariant"];
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
        AnswerSchema: {
            answer: string;
            correct: boolean;
            correction?: string;
            explanationRU?: string;
        };
        CheckBody: {
            answer: string;
            instruction: string;
            /** @enum {string} */
            kind: "merfy" | "words";
            /** Format: int64 */
            number: number;
        };
        CheckOutput: {
            correct: boolean;
            correction?: string;
            explanationRU?: string;
        };
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
        PracticeOKOutput: {
            ok: boolean;
        };
        ProgressOutput: {
            answers: {
                [key: string]: components["schemas"]["AnswerSchema"];
            };
            completed: boolean;
        };
        ReadyOutput: {
            checks?: components["schemas"]["CheckResult"][] | null;
            /** @example auth */
            module?: string;
            ready: boolean;
            /** Format: date-time */
            time: string;
        };
        SaveProgressBody: {
            answers: {
                [key: string]: components["schemas"]["AnswerSchema"];
            };
            completed: boolean;
            variant: string;
        };
        SuccessBodyCheckOutput: {
            data: components["schemas"]["CheckOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyPracticeOKOutput: {
            data: components["schemas"]["PracticeOKOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyProgressOutput: {
            data: components["schemas"]["ProgressOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyVariantSchema: {
            data: components["schemas"]["VariantSchema"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyVariantsOutput: {
            data: components["schemas"]["VariantsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        VariantSchema: {
            error?: string;
            exercises: string;
            id: string;
            /** @enum {string} */
            status: "generating" | "ready" | "error";
            title: string;
        };
        VariantsOutput: {
            items: components["schemas"]["VariantSchema"][] | null;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaAnswerSchema = components['schemas']['AnswerSchema'];
export type SchemaCheckBody = components['schemas']['CheckBody'];
export type SchemaCheckOutput = components['schemas']['CheckOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaPracticeOkOutput = components['schemas']['PracticeOKOutput'];
export type SchemaProgressOutput = components['schemas']['ProgressOutput'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSaveProgressBody = components['schemas']['SaveProgressBody'];
export type SchemaSuccessBodyCheckOutput = components['schemas']['SuccessBodyCheckOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyPracticeOkOutput = components['schemas']['SuccessBodyPracticeOKOutput'];
export type SchemaSuccessBodyProgressOutput = components['schemas']['SuccessBodyProgressOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodyVariantSchema = components['schemas']['SuccessBodyVariantSchema'];
export type SchemaSuccessBodyVariantsOutput = components['schemas']['SuccessBodyVariantsOutput'];
export type SchemaVariantSchema = components['schemas']['VariantSchema'];
export type SchemaVariantsOutput = components['schemas']['VariantsOutput'];
export type $defs = Record<string, never>;
export interface operations {
    checkPracticeAnswer: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CheckBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyCheckOutput"];
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
    deletePracticeVariant: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
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
                    "application/json": components["schemas"]["SuccessBodyPracticeOKOutput"];
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
    getPracticeProgress: {
        parameters: {
            query?: {
                variant?: string;
            };
            header?: never;
            path: {
                kind: "merfy" | "words";
                number: number;
            };
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
                    "application/json": components["schemas"]["SuccessBodyProgressOutput"];
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
    savePracticeProgress: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                kind: "merfy" | "words";
                number: number;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SaveProgressBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyPracticeOKOutput"];
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
    resetPracticeProgress: {
        parameters: {
            query?: {
                variant?: string;
            };
            header?: never;
            path: {
                kind: "merfy" | "words";
                number: number;
            };
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
                    "application/json": components["schemas"]["SuccessBodyPracticeOKOutput"];
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
    listPracticeVariants: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                kind: "merfy" | "words";
                number: number;
            };
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
                    "application/json": components["schemas"]["SuccessBodyVariantsOutput"];
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
    generatePracticeVariant: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                kind: "merfy" | "words";
                number: number;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyVariantSchema"];
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
