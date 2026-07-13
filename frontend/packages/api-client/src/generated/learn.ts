export interface paths {
    "/api/v1/books": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List available books with series, level and description */
        get: operations["listLearnBooks"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/books/{book}/chapters": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List chapters of a book */
        get: operations["listChapters"];
        put?: never;
        /** Create a chapter (global admin only) */
        post: operations["createChapter"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/books/{book}/chapters/generate": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Generate a chapter (theory + exercises) on a topic with the LLM (global admin only) */
        post: operations["generateChapter"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/books/{book}/chapters/{number}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a single chapter by number */
        get: operations["getChapter"];
        /** Update a chapter (global admin only) */
        put: operations["updateChapter"];
        post?: never;
        /** Delete a chapter (global admin only) */
        delete: operations["deleteChapter"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/illustrations": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Trigger or poll generation of an illustration from a prompt */
        post: operations["ensureIllustration"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
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
            explanation?: string;
        };
        BookListOutput: {
            items: components["schemas"]["BookSchema"][] | null;
        };
        BookSchema: {
            description?: string;
            level?: string;
            series: string;
            slug: string;
            title: string;
        };
        ChapterBody: {
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
        ChapterOutput: {
            error?: string;
            exercises: string;
            footer?: string;
            /** Format: int64 */
            number: number;
            /** Format: int64 */
            page: number;
            /** @enum {string} */
            status: "generating" | "ready" | "error";
            theory: string;
            title: string;
            words?: string[] | null;
        };
        ChaptersOutput: {
            items: components["schemas"]["ChapterOutput"][] | null;
        };
        CheckBody: {
            answer: string;
            instruction: string;
            kind: string;
            /** Format: int64 */
            number: number;
        };
        CheckOutput: {
            correct: boolean;
            correction?: string;
            explanation?: string;
        };
        CheckResult: {
            error?: string;
            /** @example postgres */
            name: string;
            ok: boolean;
        };
        DeleteChapterOutput: {
            ok: boolean;
        };
        EnsureBody: {
            /**
             * @default square
             * @enum {string}
             */
            aspect: "square" | "landscape" | "portrait";
            prompt: string;
            trigger: boolean;
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
        GenerateChapterBody: {
            topic: string;
        };
        HealthOutput: {
            /** @example auth */
            module?: string;
            /** @example ok */
            status: string;
            /** Format: date-time */
            time: string;
        };
        IllustrationOutput: {
            error?: string;
            id: string;
            /** @enum {string} */
            status: "pending" | "generating" | "ready" | "error";
            url?: string;
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
        SuccessBodyBookListOutput: {
            data: components["schemas"]["BookListOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyChapterOutput: {
            data: components["schemas"]["ChapterOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyChaptersOutput: {
            data: components["schemas"]["ChaptersOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyCheckOutput: {
            data: components["schemas"]["CheckOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyDeleteChapterOutput: {
            data: components["schemas"]["DeleteChapterOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyIllustrationOutput: {
            data: components["schemas"]["IllustrationOutput"];
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
export type SchemaBookListOutput = components['schemas']['BookListOutput'];
export type SchemaBookSchema = components['schemas']['BookSchema'];
export type SchemaChapterBody = components['schemas']['ChapterBody'];
export type SchemaChapterOutput = components['schemas']['ChapterOutput'];
export type SchemaChaptersOutput = components['schemas']['ChaptersOutput'];
export type SchemaCheckBody = components['schemas']['CheckBody'];
export type SchemaCheckOutput = components['schemas']['CheckOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaDeleteChapterOutput = components['schemas']['DeleteChapterOutput'];
export type SchemaEnsureBody = components['schemas']['EnsureBody'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaGenerateChapterBody = components['schemas']['GenerateChapterBody'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaIllustrationOutput = components['schemas']['IllustrationOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaPracticeOkOutput = components['schemas']['PracticeOKOutput'];
export type SchemaProgressOutput = components['schemas']['ProgressOutput'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSaveProgressBody = components['schemas']['SaveProgressBody'];
export type SchemaSuccessBodyBookListOutput = components['schemas']['SuccessBodyBookListOutput'];
export type SchemaSuccessBodyChapterOutput = components['schemas']['SuccessBodyChapterOutput'];
export type SchemaSuccessBodyChaptersOutput = components['schemas']['SuccessBodyChaptersOutput'];
export type SchemaSuccessBodyCheckOutput = components['schemas']['SuccessBodyCheckOutput'];
export type SchemaSuccessBodyDeleteChapterOutput = components['schemas']['SuccessBodyDeleteChapterOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyIllustrationOutput = components['schemas']['SuccessBodyIllustrationOutput'];
export type SchemaSuccessBodyPracticeOkOutput = components['schemas']['SuccessBodyPracticeOKOutput'];
export type SchemaSuccessBodyProgressOutput = components['schemas']['SuccessBodyProgressOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodyVariantSchema = components['schemas']['SuccessBodyVariantSchema'];
export type SchemaSuccessBodyVariantsOutput = components['schemas']['SuccessBodyVariantsOutput'];
export type SchemaVariantSchema = components['schemas']['VariantSchema'];
export type SchemaVariantsOutput = components['schemas']['VariantsOutput'];
export type $defs = Record<string, never>;
export interface operations {
    listLearnBooks: {
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
                    "application/json": components["schemas"]["SuccessBodyBookListOutput"];
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
    listChapters: {
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
                    "application/json": components["schemas"]["SuccessBodyChaptersOutput"];
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
    createChapter: {
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
                "application/json": components["schemas"]["ChapterBody"];
            };
        };
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyChapterOutput"];
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
    generateChapter: {
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
                "application/json": components["schemas"]["GenerateChapterBody"];
            };
        };
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyChapterOutput"];
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
    getChapter: {
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
                    "application/json": components["schemas"]["SuccessBodyChapterOutput"];
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
    updateChapter: {
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
                "application/json": components["schemas"]["ChapterBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyChapterOutput"];
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
    deleteChapter: {
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
                    "application/json": components["schemas"]["SuccessBodyDeleteChapterOutput"];
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
    ensureIllustration: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["EnsureBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyIllustrationOutput"];
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
                kind: string;
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
                kind: string;
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
                kind: string;
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
                kind: string;
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
                kind: string;
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
