export interface paths {
    "/api/v1/books/{book}/lessons": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List lessons of a book */
        get: operations["listBookLessons"];
        put?: never;
        /** Create a lesson (global admin only) */
        post: operations["createBookLesson"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/books/{book}/lessons/{number}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a single lesson by number */
        get: operations["getBookLesson"];
        /** Update a lesson (global admin only) */
        put: operations["updateBookLesson"];
        post?: never;
        /** Delete a lesson (global admin only) */
        delete: operations["deleteBookLesson"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        CheckResult: {
            error?: string;
            /** @example postgres */
            name: string;
            ok: boolean;
        };
        DeleteLessonOutput: {
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
        LessonBody: {
            exercises: string;
            footer: string;
            /** Format: int64 */
            number: number;
            /** Format: int64 */
            page: number;
            theory: string;
            title: string;
            words: string[] | null;
        };
        LessonOutput: {
            exercises: string;
            footer?: string;
            /** Format: int64 */
            number: number;
            /** Format: int64 */
            page: number;
            theory: string;
            title: string;
            words: string[] | null;
        };
        LessonsOutput: {
            items: components["schemas"]["LessonOutput"][] | null;
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
        SuccessBodyDeleteLessonOutput: {
            data: components["schemas"]["DeleteLessonOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyLessonOutput: {
            data: components["schemas"]["LessonOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyLessonsOutput: {
            data: components["schemas"]["LessonsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaDeleteLessonOutput = components['schemas']['DeleteLessonOutput'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaLessonBody = components['schemas']['LessonBody'];
export type SchemaLessonOutput = components['schemas']['LessonOutput'];
export type SchemaLessonsOutput = components['schemas']['LessonsOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSuccessBodyDeleteLessonOutput = components['schemas']['SuccessBodyDeleteLessonOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyLessonOutput = components['schemas']['SuccessBodyLessonOutput'];
export type SchemaSuccessBodyLessonsOutput = components['schemas']['SuccessBodyLessonsOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type $defs = Record<string, never>;
export interface operations {
    listBookLessons: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                book: string;
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
                    "application/json": components["schemas"]["SuccessBodyLessonsOutput"];
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
    createBookLesson: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                book: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["LessonBody"];
            };
        };
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyLessonOutput"];
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
    getBookLesson: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                book: string;
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
                    "application/json": components["schemas"]["SuccessBodyLessonOutput"];
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
    updateBookLesson: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                book: string;
                number: number;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["LessonBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyLessonOutput"];
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
    deleteBookLesson: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                book: string;
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
                    "application/json": components["schemas"]["SuccessBodyDeleteLessonOutput"];
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
