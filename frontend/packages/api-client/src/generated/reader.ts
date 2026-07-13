export interface paths {
    "/api/v1/reader/articles": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Import an article by URL; readable text and images are extracted and converted in the background */
        post: operations["importArticle"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/reader/books": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List the reader's books */
        get: operations["listBooks"];
        put?: never;
        /** Upload a book (FB2/EPUB/HTML); it is converted to HTML in the background */
        post: operations["uploadBook"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/reader/books/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a book with its content URL and reading position */
        get: operations["getBook"];
        put?: never;
        post?: never;
        /** Delete a book */
        delete: operations["deleteBook"];
        options?: never;
        head?: never;
        /** Update a book's title, author, description and cover */
        patch: operations["updateBook"];
        trace?: never;
    };
    "/api/v1/reader/books/{id}/progress": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Save the reading position for a book */
        put: operations["saveBookProgress"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/reader/collections": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List article collections (cover and description) */
        get: operations["listCollections"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** Update an article collection's title, description and cover */
        patch: operations["updateCollection"];
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        BookOutput: {
            author?: string;
            content_url?: string;
            cover_url?: string;
            created_at: string;
            description?: string;
            error?: string;
            group_title?: string;
            id: string;
            kind: string;
            /** Format: int64 */
            percent: number;
            /** Format: int64 */
            position: number;
            status: string;
            /** Format: int64 */
            text_length: number;
            title: string;
        };
        BookSummary: {
            author?: string;
            cover_url?: string;
            created_at: string;
            description?: string;
            group_title?: string;
            id: string;
            kind: string;
            /** Format: int64 */
            percent: number;
            /** Format: int64 */
            position: number;
            status: string;
            /** Format: int64 */
            text_length: number;
            title: string;
        };
        BooksOutput: {
            items: components["schemas"]["BookSummary"][] | null;
        };
        CheckResult: {
            error?: string;
            /** @example postgres */
            name: string;
            ok: boolean;
        };
        CollectionOutput: {
            cover_url?: string;
            description?: string;
            title: string;
        };
        CollectionsOutput: {
            items: components["schemas"]["CollectionOutput"][] | null;
        };
        DeleteBookOutput: {
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
        FormFile: {
            ContentType: string;
            Filename: string;
            IsSet: boolean;
            /** Format: int64 */
            Size: number;
        };
        HealthOutput: {
            /** @example auth */
            module?: string;
            /** @example ok */
            status: string;
            /** Format: date-time */
            time: string;
        };
        ImportArticleInputBody: {
            /** @description Optional collection to group the article under */
            group_title?: string;
            /**
             * Format: uri
             * @description Public article URL
             */
            url: string;
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
        SaveBookProgressInputBody: {
            /** Format: int64 */
            position: number;
        };
        SaveBookProgressOutput: {
            ok: boolean;
        };
        SuccessBodyBookOutput: {
            data: components["schemas"]["BookOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyBooksOutput: {
            data: components["schemas"]["BooksOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyCollectionOutput: {
            data: components["schemas"]["CollectionOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyCollectionsOutput: {
            data: components["schemas"]["CollectionsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyDeleteBookOutput: {
            data: components["schemas"]["DeleteBookOutput"];
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
        SuccessBodySaveBookProgressOutput: {
            data: components["schemas"]["SaveBookProgressOutput"];
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
export type SchemaBookOutput = components['schemas']['BookOutput'];
export type SchemaBookSummary = components['schemas']['BookSummary'];
export type SchemaBooksOutput = components['schemas']['BooksOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaCollectionOutput = components['schemas']['CollectionOutput'];
export type SchemaCollectionsOutput = components['schemas']['CollectionsOutput'];
export type SchemaDeleteBookOutput = components['schemas']['DeleteBookOutput'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaFormFile = components['schemas']['FormFile'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaImportArticleInputBody = components['schemas']['ImportArticleInputBody'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSaveBookProgressInputBody = components['schemas']['SaveBookProgressInputBody'];
export type SchemaSaveBookProgressOutput = components['schemas']['SaveBookProgressOutput'];
export type SchemaSuccessBodyBookOutput = components['schemas']['SuccessBodyBookOutput'];
export type SchemaSuccessBodyBooksOutput = components['schemas']['SuccessBodyBooksOutput'];
export type SchemaSuccessBodyCollectionOutput = components['schemas']['SuccessBodyCollectionOutput'];
export type SchemaSuccessBodyCollectionsOutput = components['schemas']['SuccessBodyCollectionsOutput'];
export type SchemaSuccessBodyDeleteBookOutput = components['schemas']['SuccessBodyDeleteBookOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodySaveBookProgressOutput = components['schemas']['SuccessBodySaveBookProgressOutput'];
export type $defs = Record<string, never>;
export interface operations {
    importArticle: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ImportArticleInputBody"];
            };
        };
        responses: {
            /** @description Accepted */
            202: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyBookOutput"];
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
    listBooks: {
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
                    "application/json": components["schemas"]["SuccessBodyBooksOutput"];
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
    uploadBook: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "multipart/form-data": {
                    /** Format: binary */
                    cover?: string;
                    /** Format: binary */
                    file: string;
                };
            };
        };
        responses: {
            /** @description Accepted */
            202: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyBookOutput"];
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
    getBook: {
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
                    "application/json": components["schemas"]["SuccessBodyBookOutput"];
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
    deleteBook: {
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
                    "application/json": components["schemas"]["SuccessBodyDeleteBookOutput"];
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
    updateBook: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: {
            content: {
                "multipart/form-data": {
                    /** Format: binary */
                    cover?: string;
                };
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyBookOutput"];
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
    saveBookProgress: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SaveBookProgressInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodySaveBookProgressOutput"];
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
    listCollections: {
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
                    "application/json": components["schemas"]["SuccessBodyCollectionsOutput"];
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
    updateCollection: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "multipart/form-data": {
                    /** Format: binary */
                    cover?: string;
                };
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyCollectionOutput"];
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
