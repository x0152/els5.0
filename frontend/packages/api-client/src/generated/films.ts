export interface paths {
    "/api/v1/films": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List films */
        get: operations["listFilms"];
        put?: never;
        /** Upload a film; tracks are transcoded in the background (global admin only) */
        post: operations["uploadFilm"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/films/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a film with audio and subtitle tracks */
        get: operations["getFilm"];
        put?: never;
        post?: never;
        /** Delete a film (global admin only) */
        delete: operations["deleteFilm"];
        options?: never;
        head?: never;
        /** Update a film's title, description and poster (global admin only) */
        patch: operations["updateFilm"];
        trace?: never;
    };
    "/api/v1/films/{id}/progress": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Save the watch position for a film */
        put: operations["saveFilmProgress"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/series": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List series metadata (cover and description) */
        get: operations["listSeries"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** Update a series title, description and cover (global admin only) */
        patch: operations["updateSeries"];
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        AudioTrackOutput: {
            label: string;
            lang: string;
            url: string;
        };
        CheckResult: {
            error?: string;
            /** @example postgres */
            name: string;
            ok: boolean;
        };
        CueOutput: {
            /** Format: int64 */
            end_ms: number;
            /** Format: int64 */
            index: number;
            /** Format: int64 */
            start_ms: number;
            text: string;
        };
        DeleteFilmOutput: {
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
        FilmOutput: {
            audio_tracks: components["schemas"]["AudioTrackOutput"][] | null;
            created_at: string;
            description?: string;
            /** Format: int64 */
            duration_ms: number;
            /** Format: int64 */
            episode: number;
            error?: string;
            id: string;
            kind: string;
            level: string;
            /** Format: int64 */
            position_ms: number;
            poster_url?: string;
            /** Format: int64 */
            season: number;
            series_title?: string;
            status: string;
            subtitles: components["schemas"]["SubtitleTrackOutput"][] | null;
            title: string;
        };
        FilmSummary: {
            created_at: string;
            description?: string;
            /** Format: int64 */
            duration_ms: number;
            /** Format: int64 */
            episode: number;
            id: string;
            kind: string;
            level: string;
            /** Format: int64 */
            position_ms: number;
            poster_url?: string;
            /** Format: int64 */
            season: number;
            series_title?: string;
            status: string;
            title: string;
        };
        FilmsOutput: {
            items: components["schemas"]["FilmSummary"][] | null;
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
        SaveProgressInputBody: {
            /** Format: int64 */
            position_ms: number;
        };
        SaveProgressOutput: {
            ok: boolean;
        };
        SeriesListOutput: {
            items: components["schemas"]["SeriesOutput"][] | null;
        };
        SeriesOutput: {
            description?: string;
            poster_url?: string;
            title: string;
        };
        SubtitleTrackOutput: {
            cues: components["schemas"]["CueOutput"][] | null;
            label: string;
            lang: string;
        };
        SuccessBodyDeleteFilmOutput: {
            data: components["schemas"]["DeleteFilmOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyFilmOutput: {
            data: components["schemas"]["FilmOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyFilmsOutput: {
            data: components["schemas"]["FilmsOutput"];
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
        SuccessBodySaveProgressOutput: {
            data: components["schemas"]["SaveProgressOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodySeriesListOutput: {
            data: components["schemas"]["SeriesListOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodySeriesOutput: {
            data: components["schemas"]["SeriesOutput"];
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
export type SchemaAudioTrackOutput = components['schemas']['AudioTrackOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaCueOutput = components['schemas']['CueOutput'];
export type SchemaDeleteFilmOutput = components['schemas']['DeleteFilmOutput'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaFilmOutput = components['schemas']['FilmOutput'];
export type SchemaFilmSummary = components['schemas']['FilmSummary'];
export type SchemaFilmsOutput = components['schemas']['FilmsOutput'];
export type SchemaFormFile = components['schemas']['FormFile'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSaveProgressInputBody = components['schemas']['SaveProgressInputBody'];
export type SchemaSaveProgressOutput = components['schemas']['SaveProgressOutput'];
export type SchemaSeriesListOutput = components['schemas']['SeriesListOutput'];
export type SchemaSeriesOutput = components['schemas']['SeriesOutput'];
export type SchemaSubtitleTrackOutput = components['schemas']['SubtitleTrackOutput'];
export type SchemaSuccessBodyDeleteFilmOutput = components['schemas']['SuccessBodyDeleteFilmOutput'];
export type SchemaSuccessBodyFilmOutput = components['schemas']['SuccessBodyFilmOutput'];
export type SchemaSuccessBodyFilmsOutput = components['schemas']['SuccessBodyFilmsOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodySaveProgressOutput = components['schemas']['SuccessBodySaveProgressOutput'];
export type SchemaSuccessBodySeriesListOutput = components['schemas']['SuccessBodySeriesListOutput'];
export type SchemaSuccessBodySeriesOutput = components['schemas']['SuccessBodySeriesOutput'];
export type $defs = Record<string, never>;
export interface operations {
    listFilms: {
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
                    "application/json": components["schemas"]["SuccessBodyFilmsOutput"];
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
    uploadFilm: {
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
                    poster?: string;
                    /** Format: binary */
                    subtitles?: string;
                    /** Format: binary */
                    video: string;
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
                    "application/json": components["schemas"]["SuccessBodyFilmOutput"];
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
    getFilm: {
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
                    "application/json": components["schemas"]["SuccessBodyFilmOutput"];
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
    deleteFilm: {
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
                    "application/json": components["schemas"]["SuccessBodyDeleteFilmOutput"];
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
    updateFilm: {
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
                    poster?: string;
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
                    "application/json": components["schemas"]["SuccessBodyFilmOutput"];
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
    saveFilmProgress: {
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
                "application/json": components["schemas"]["SaveProgressInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodySaveProgressOutput"];
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
    listSeries: {
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
                    "application/json": components["schemas"]["SuccessBodySeriesListOutput"];
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
    updateSeries: {
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
                    poster?: string;
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
                    "application/json": components["schemas"]["SuccessBodySeriesOutput"];
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
