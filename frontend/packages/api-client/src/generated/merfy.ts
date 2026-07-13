export interface paths {
    "/api/v1/merfy/units": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List merfy units */
        get: operations["listMerfyUnits"];
        put?: never;
        /** Create a merfy unit (global admin only) */
        post: operations["createMerfyUnit"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/merfy/units/{number}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a single merfy unit by number */
        get: operations["getMerfyUnit"];
        /** Update a merfy unit (global admin only) */
        put: operations["updateMerfyUnit"];
        post?: never;
        /** Delete a merfy unit (global admin only) */
        delete: operations["deleteMerfyUnit"];
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
        DeleteUnitOutput: {
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
        ReadyOutput: {
            checks?: components["schemas"]["CheckResult"][] | null;
            /** @example auth */
            module?: string;
            ready: boolean;
            /** Format: date-time */
            time: string;
        };
        SuccessBodyDeleteUnitOutput: {
            data: components["schemas"]["DeleteUnitOutput"];
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
        SuccessBodyUnitOutput: {
            data: components["schemas"]["UnitOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyUnitsOutput: {
            data: components["schemas"]["UnitsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        UnitBody: {
            exercises: string;
            footer: string;
            /** Format: int64 */
            number: number;
            /** Format: int64 */
            page: number;
            theory: string;
            title: string;
        };
        UnitOutput: {
            exercises: string;
            footer?: string;
            /** Format: int64 */
            number: number;
            /** Format: int64 */
            page: number;
            theory: string;
            title: string;
        };
        UnitsOutput: {
            items: components["schemas"]["UnitOutput"][] | null;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaDeleteUnitOutput = components['schemas']['DeleteUnitOutput'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSuccessBodyDeleteUnitOutput = components['schemas']['SuccessBodyDeleteUnitOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodyUnitOutput = components['schemas']['SuccessBodyUnitOutput'];
export type SchemaSuccessBodyUnitsOutput = components['schemas']['SuccessBodyUnitsOutput'];
export type SchemaUnitBody = components['schemas']['UnitBody'];
export type SchemaUnitOutput = components['schemas']['UnitOutput'];
export type SchemaUnitsOutput = components['schemas']['UnitsOutput'];
export type $defs = Record<string, never>;
export interface operations {
    listMerfyUnits: {
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
                    "application/json": components["schemas"]["SuccessBodyUnitsOutput"];
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
    createMerfyUnit: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UnitBody"];
            };
        };
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyUnitOutput"];
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
    getMerfyUnit: {
        parameters: {
            query?: never;
            header?: never;
            path: {
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
                    "application/json": components["schemas"]["SuccessBodyUnitOutput"];
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
    updateMerfyUnit: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                number: number;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UnitBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyUnitOutput"];
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
    deleteMerfyUnit: {
        parameters: {
            query?: never;
            header?: never;
            path: {
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
                    "application/json": components["schemas"]["SuccessBodyDeleteUnitOutput"];
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
