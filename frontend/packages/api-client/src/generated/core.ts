export interface paths {
    "/api/v1/core/catalog": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List word and grammar catalog */
        get: operations["listCoreCatalog"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/core/data": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Wipe own events */
        delete: operations["wipeCore"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/core/dictionaries": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List column dictionaries */
        get: operations["listCoreDictionaries"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/core/events": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List learning events by status */
        get: operations["listCoreEvents"];
        put?: never;
        /** Ingest learning events */
        post: operations["ingestCoreEvents"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/core/events/unclear": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Mark a heard line as not understood */
        post: operations["markCoreEventUnclear"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/core/rows": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Delete selected rows */
        delete: operations["deleteCoreRows"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        CatalogOutput: {
            rules: components["schemas"]["GrammarRuleView"][] | null;
            words: components["schemas"]["WordView"][] | null;
        };
        CheckResult: {
            error?: string;
            /** @example postgres */
            name: string;
            ok: boolean;
        };
        DeleteRowsInputBody: {
            ids: string[] | null;
            /** @enum {string} */
            kind: "events" | "raw" | "words" | "rules";
        };
        DeleteRowsOutput: {
            /** Format: int64 */
            deleted: number;
        };
        DictEntryView: {
            icon?: string;
            label: string;
            value: string;
        };
        DictionariesOutput: {
            dictionaries: {
                [key: string]: components["schemas"]["DictEntryView"][] | null;
            };
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
        EventEnvelope: {
            client_id?: string;
            context?: string;
            /** @description free bag, e.g. precise internal keys */
            meta?: {
                [key: string]: unknown;
            };
            /** Format: date-time */
            occurred_at?: string;
            /** @description ok|fail (for targeted input) */
            outcome?: string;
            /** @description reading|writing|speaking|listening (optional for pure self-assessment) */
            skill?: string;
            /** @description provenance: app, book_id, video_id, ... */
            source?: {
                [key: string]: unknown;
            };
            /** @description word or grammar concept practiced, plain language (targeted input) */
            target?: string;
            /** @description language sample to decompose (free input) */
            text?: string;
        };
        EventView: {
            action?: string;
            client_id?: string;
            context?: string;
            /** Format: date-time */
            created_at: string;
            error?: {
                [key: string]: unknown;
            };
            grammar_key?: string;
            id: string;
            lemma?: string;
            meta?: {
                [key: string]: unknown;
            };
            /** Format: date-time */
            occurred_at: string;
            outcome?: string;
            pos?: string;
            raw_event_id?: string;
            skill?: string;
            source?: {
                [key: string]: unknown;
            };
            status: string;
            target?: string;
            text?: string;
        };
        GrammarRuleView: {
            cefr_level?: string;
            /** Format: date-time */
            created_at: string;
            enriched: boolean;
            id: string;
            key: string;
            parent_key?: string;
            title?: string;
        };
        HealthOutput: {
            /** @example auth */
            module?: string;
            /** @example ok */
            status: string;
            /** Format: date-time */
            time: string;
        };
        IngestInputBody: {
            events: components["schemas"]["EventEnvelope"][] | null;
        };
        IngestOutput: {
            /** Format: int64 */
            accepted: number;
        };
        ListOutput: {
            events: components["schemas"]["EventView"][] | null;
        };
        MarkUnclearOutput: {
            updated: boolean;
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
        SuccessBodyCatalogOutput: {
            data: components["schemas"]["CatalogOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyDeleteRowsOutput: {
            data: components["schemas"]["DeleteRowsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyDictionariesOutput: {
            data: components["schemas"]["DictionariesOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyIngestOutput: {
            data: components["schemas"]["IngestOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyListOutput: {
            data: components["schemas"]["ListOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyMarkUnclearOutput: {
            data: components["schemas"]["MarkUnclearOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyWipeOutput: {
            data: components["schemas"]["WipeOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        WipeOutput: {
            ok: boolean;
        };
        WordView: {
            cefr?: string;
            /** Format: date-time */
            created_at: string;
            enriched: boolean;
            /** Format: double */
            frequency?: number;
            id: string;
            key: string;
            lemma: string;
            pos?: string;
            type?: string;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaCatalogOutput = components['schemas']['CatalogOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaDeleteRowsInputBody = components['schemas']['DeleteRowsInputBody'];
export type SchemaDeleteRowsOutput = components['schemas']['DeleteRowsOutput'];
export type SchemaDictEntryView = components['schemas']['DictEntryView'];
export type SchemaDictionariesOutput = components['schemas']['DictionariesOutput'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaEventEnvelope = components['schemas']['EventEnvelope'];
export type SchemaEventView = components['schemas']['EventView'];
export type SchemaGrammarRuleView = components['schemas']['GrammarRuleView'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaIngestInputBody = components['schemas']['IngestInputBody'];
export type SchemaIngestOutput = components['schemas']['IngestOutput'];
export type SchemaListOutput = components['schemas']['ListOutput'];
export type SchemaMarkUnclearOutput = components['schemas']['MarkUnclearOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSuccessBodyCatalogOutput = components['schemas']['SuccessBodyCatalogOutput'];
export type SchemaSuccessBodyDeleteRowsOutput = components['schemas']['SuccessBodyDeleteRowsOutput'];
export type SchemaSuccessBodyDictionariesOutput = components['schemas']['SuccessBodyDictionariesOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyIngestOutput = components['schemas']['SuccessBodyIngestOutput'];
export type SchemaSuccessBodyListOutput = components['schemas']['SuccessBodyListOutput'];
export type SchemaSuccessBodyMarkUnclearOutput = components['schemas']['SuccessBodyMarkUnclearOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodyWipeOutput = components['schemas']['SuccessBodyWipeOutput'];
export type SchemaWipeOutput = components['schemas']['WipeOutput'];
export type SchemaWordView = components['schemas']['WordView'];
export type $defs = Record<string, never>;
export interface operations {
    listCoreCatalog: {
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
                    "application/json": components["schemas"]["SuccessBodyCatalogOutput"];
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
    wipeCore: {
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
                    "application/json": components["schemas"]["SuccessBodyWipeOutput"];
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
    listCoreDictionaries: {
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
                    "application/json": components["schemas"]["SuccessBodyDictionariesOutput"];
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
    listCoreEvents: {
        parameters: {
            query?: {
                status?: "pending" | "processed" | "failed" | "all" | "raw";
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
                    "application/json": components["schemas"]["SuccessBodyListOutput"];
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
    ingestCoreEvents: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["IngestInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyIngestOutput"];
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
    markCoreEventUnclear: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["EventEnvelope"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyMarkUnclearOutput"];
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
    deleteCoreRows: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DeleteRowsInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyDeleteRowsOutput"];
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
