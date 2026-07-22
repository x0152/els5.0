export interface paths {
    "/api/v1/workout": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Delete all workout progress and generated lessons of the account */
        delete: operations["workoutReset"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/workout/lessons": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Return the active lesson or start background generation */
        post: operations["workoutStartLesson"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/workout/lessons/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a lesson with all step payloads */
        get: operations["workoutGetLesson"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/workout/lessons/{id}/steps/{step}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Submit a step result; the lesson completes with its last step */
        post: operations["workoutSubmitStep"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/workout/today": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Current lesson, streak and completion calendar */
        get: operations["workoutToday"];
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
        GrammarOutput: {
            exercises: string;
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
        ItemResultInput: {
            /** Format: int64 */
            end_ms?: number;
            film_id?: string;
            /** @enum {string} */
            kind: "phrase" | "word";
            /** Format: int64 */
            score: number;
            /** Format: int64 */
            start_ms?: number;
            text: string;
        };
        LessonOutput: {
            created_at: string;
            /** Format: int64 */
            cycle_index: number;
            /** Format: int64 */
            end_ms: number;
            film_id?: string;
            id: string;
            /** Format: int64 */
            number: number;
            review: boolean;
            /** Format: int64 */
            start_ms: number;
            /** @enum {string} */
            status: "active" | "completed";
            steps: components["schemas"]["StepOutput"][] | null;
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
        PhraseOutput: {
            /** Format: int64 */
            end_ms?: number;
            film_id?: string;
            /** Format: int64 */
            start_ms?: number;
            text: string;
        };
        QuestionOutput: {
            /** Format: int64 */
            answer: number;
            options: string[] | null;
            text: string;
        };
        ReadingOutput: {
            body: string;
            title: string;
            words?: string[] | null;
        };
        ReadyOutput: {
            checks?: components["schemas"]["CheckResult"][] | null;
            /** @example auth */
            module?: string;
            ready: boolean;
            /** Format: date-time */
            time: string;
        };
        ResetOutput: Record<string, never>;
        StartLessonOutput: {
            generating: boolean;
            lesson?: components["schemas"]["LessonOutput"];
        };
        StepOutput: {
            done: boolean;
            grammar?: components["schemas"]["GrammarOutput"];
            id: string;
            /** @enum {string} */
            kind: "warmup" | "watch" | "questions" | "speak" | "dictation" | "reading" | "writing" | "grammar" | "vocab";
            phrases?: components["schemas"]["PhraseOutput"][] | null;
            questions?: components["schemas"]["QuestionOutput"][] | null;
            reading?: components["schemas"]["ReadingOutput"];
            /** Format: int64 */
            score: number;
            title: string;
            vocab?: components["schemas"]["VocabWordOutput"][] | null;
            warmup?: components["schemas"]["WarmupItemOutput"][] | null;
            watch?: components["schemas"]["WatchOutput"];
            writing?: components["schemas"]["WritingOutput"];
        };
        SubmitStepInputBody: {
            results?: components["schemas"]["ItemResultInput"][] | null;
            /** Format: int64 */
            score: number;
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
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyResetOutput: {
            data: components["schemas"]["ResetOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyStartLessonOutput: {
            data: components["schemas"]["StartLessonOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyWorkoutTodayOutput: {
            data: components["schemas"]["WorkoutTodayOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        VocabWordOutput: {
            definition?: string;
            example?: string;
            text: string;
            translation?: string;
        };
        WarmupItemOutput: {
            /** Format: int64 */
            end_ms?: number;
            film_id?: string;
            /** @enum {string} */
            mode: "speak" | "dictation";
            /** Format: int64 */
            start_ms?: number;
            text: string;
        };
        WatchOutput: {
            /** Format: int64 */
            end_ms: number;
            film_id: string;
            recap?: string;
            /** Format: int64 */
            start_ms: number;
            summary?: string;
            title: string;
        };
        WorkoutTodayOutput: {
            completed: boolean;
            days: string[] | null;
            generating?: boolean;
            generating_since?: string;
            generation_failed?: boolean;
            lesson?: components["schemas"]["LessonOutput"];
            /** Format: int64 */
            streak: number;
        };
        WritingOutput: {
            dialogue: string;
            scenario: string;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaGrammarOutput = components['schemas']['GrammarOutput'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaItemResultInput = components['schemas']['ItemResultInput'];
export type SchemaLessonOutput = components['schemas']['LessonOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaPhraseOutput = components['schemas']['PhraseOutput'];
export type SchemaQuestionOutput = components['schemas']['QuestionOutput'];
export type SchemaReadingOutput = components['schemas']['ReadingOutput'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaResetOutput = components['schemas']['ResetOutput'];
export type SchemaStartLessonOutput = components['schemas']['StartLessonOutput'];
export type SchemaStepOutput = components['schemas']['StepOutput'];
export type SchemaSubmitStepInputBody = components['schemas']['SubmitStepInputBody'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyLessonOutput = components['schemas']['SuccessBodyLessonOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodyResetOutput = components['schemas']['SuccessBodyResetOutput'];
export type SchemaSuccessBodyStartLessonOutput = components['schemas']['SuccessBodyStartLessonOutput'];
export type SchemaSuccessBodyWorkoutTodayOutput = components['schemas']['SuccessBodyWorkoutTodayOutput'];
export type SchemaVocabWordOutput = components['schemas']['VocabWordOutput'];
export type SchemaWarmupItemOutput = components['schemas']['WarmupItemOutput'];
export type SchemaWatchOutput = components['schemas']['WatchOutput'];
export type SchemaWorkoutTodayOutput = components['schemas']['WorkoutTodayOutput'];
export type SchemaWritingOutput = components['schemas']['WritingOutput'];
export type $defs = Record<string, never>;
export interface operations {
    workoutReset: {
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
                    "application/json": components["schemas"]["SuccessBodyResetOutput"];
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
    workoutStartLesson: {
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
                    "application/json": components["schemas"]["SuccessBodyStartLessonOutput"];
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
    workoutGetLesson: {
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
    workoutSubmitStep: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
                step: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SubmitStepInputBody"];
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
    workoutToday: {
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
                    "application/json": components["schemas"]["SuccessBodyWorkoutTodayOutput"];
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
