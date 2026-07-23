export interface paths {
    "/api/v1/vocab/analyze": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Break selected text into candidate vocabulary items via the LLM */
        post: operations["analyzeVocab"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/cards": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Build a flashcard deck: guess the word by its image and definition */
        post: operations["generateVocabCards"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/cards/answer": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Check a flashcard answer and advance the word's memorization progress */
        post: operations["answerVocabCard"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/cards/due": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Count words that can still advance their memorization progress today */
        get: operations["dueVocabCards"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/occurrences": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List media where a word or phrase occurs, from the parsed lexicon */
        get: operations["vocabOccurrences"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/practice": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get the learner's latest practice session, with saved answers */
        get: operations["getVocabPractice"];
        put?: never;
        /** Generate a fresh LLM practice session from the learner's learning words */
        post: operations["generateVocabPractice"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/practice/check": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Check a free-form practice sentence via the LLM */
        post: operations["checkVocabPractice"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/practice/progress": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** Save the learner's answers for the current practice session */
        put: operations["saveVocabPracticeProgress"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/units": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List vocabulary items with search, status filter and pagination */
        get: operations["listVocabUnits"];
        put?: never;
        /** Add a vocabulary item; the LLM validates it and writes its description */
        post: operations["addVocabUnit"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/units/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Delete a vocabulary item */
        delete: operations["deleteVocabUnit"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/vocab/units/{id}/status": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** Update a vocabulary item's memorization status */
        patch: operations["updateVocabUnitStatus"];
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        AddUnitInputBody: {
            /** @description Known CEFR level from analyze */
            cefr?: string;
            /** @description Known short definition from analyze */
            description?: string;
            /**
             * Format: int64
             * @description Known frequency from analyze
             */
            frequency?: number;
            /** @description Known kind from analyze; when set the unit is stored instantly and enriched in the background */
            kind?: string;
            /** @description Word, phrase, phrasal verb or idiom to add */
            text: string;
            /** @description Known translation from analyze */
            translation?: string;
        };
        AddUnitOutput: {
            correct: boolean;
            correction?: string;
            explanation?: string;
            unit?: components["schemas"]["UnitOutput"];
        };
        AnalyzeInputBody: {
            /** @description Optional surrounding text for disambiguation */
            context?: string;
            /** @description Selected text to break into vocabulary items */
            text: string;
        };
        AnalyzeItemOutput: {
            cefr: string;
            common: boolean;
            description: string;
            existing: boolean;
            /** Format: int64 */
            frequency: number;
            kind: string;
            media: components["schemas"]["OccurrenceOutput"][] | null;
            /** Format: int64 */
            media_count: number;
            text: string;
            /** Format: int64 */
            total: number;
            translation?: string;
        };
        AnalyzeOutput: {
            items: components["schemas"]["AnalyzeItemOutput"][] | null;
        };
        AnswerCardInputBody: {
            answer: string;
            unit_id: string;
        };
        AnswerCardOutput: {
            correct: boolean;
            unit: components["schemas"]["UnitOutput"];
        };
        CardOutput: {
            definition?: string;
            /** @enum {string} */
            direction: "word" | "translation";
            image_url?: string;
            kind: string;
            /** @enum {string} */
            mode: "choice" | "input";
            options?: string[] | null;
            status: string;
            transcription?: string;
            unit_id: string;
            word?: string;
        };
        CardsOutput: {
            cards: components["schemas"]["CardOutput"][] | null;
        };
        CheckPracticeInputBody: {
            /** @description The learner's free-form sentence */
            answer: string;
            /** @description The exercise instruction the answer responds to */
            instruction: string;
        };
        CheckPracticeOutput: {
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
        DeleteUnitOutput: {
            ok: boolean;
        };
        DueCardsOutput: {
            /** Format: int64 */
            count: number;
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
        GenerateCardsInputBody: {
            /** @description Build the deck only from words with a generated illustration */
            images_only?: boolean;
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
        OccurrenceOutput: {
            author?: string;
            /** Format: int64 */
            count: number;
            /** Format: int64 */
            episode?: number;
            kind?: string;
            media_id: string;
            media_type: string;
            /** Format: int64 */
            season?: number;
            series_title?: string;
            spots: components["schemas"]["SpotOutput"][] | null;
            title: string;
        };
        OccurrencesOutput: {
            common: boolean;
            media: components["schemas"]["OccurrenceOutput"][] | null;
            /** Format: int64 */
            media_count: number;
            /** Format: int64 */
            total: number;
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
        PracticeAnswerDTO: {
            answer: string;
            correct: boolean;
            correction?: string;
            explanation?: string;
        };
        PracticeOutput: {
            answers: {
                [key: string]: components["schemas"]["PracticeAnswerDTO"];
            };
            completed: boolean;
            error?: string;
            exercises: string;
            id: string;
            status: string;
            words: components["schemas"]["UnitOutput"][] | null;
        };
        ReadyOutput: {
            checks?: components["schemas"]["CheckResult"][] | null;
            /** @example auth */
            module?: string;
            ready: boolean;
            /** Format: date-time */
            time: string;
        };
        SavePracticeProgressInputBody: {
            answers: {
                [key: string]: components["schemas"]["PracticeAnswerDTO"];
            };
            completed: boolean;
            session_id: string;
        };
        SavePracticeProgressOutput: {
            ok: boolean;
        };
        SpotOutput: {
            example?: string;
            /** Format: int64 */
            ref: number;
        };
        SuccessBodyAddUnitOutput: {
            data: components["schemas"]["AddUnitOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyAnalyzeOutput: {
            data: components["schemas"]["AnalyzeOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyAnswerCardOutput: {
            data: components["schemas"]["AnswerCardOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyCardsOutput: {
            data: components["schemas"]["CardsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyCheckPracticeOutput: {
            data: components["schemas"]["CheckPracticeOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyDeleteUnitOutput: {
            data: components["schemas"]["DeleteUnitOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyDueCardsOutput: {
            data: components["schemas"]["DueCardsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyOccurrencesOutput: {
            data: components["schemas"]["OccurrencesOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyPracticeOutput: {
            data: components["schemas"]["PracticeOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodySavePracticeProgressOutput: {
            data: components["schemas"]["SavePracticeProgressOutput"];
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
        UnitOutput: {
            cefr: string;
            /** Format: int64 */
            correct_streak: number;
            created_at: string;
            definition?: string;
            example?: string;
            /** Format: int64 */
            frequency: number;
            id: string;
            kind: string;
            status: string;
            text: string;
            transcription?: string;
            translation?: string;
        };
        UnitsOutput: {
            items: components["schemas"]["UnitOutput"][] | null;
            /** Format: int64 */
            limit: number;
            /** Format: int64 */
            offset: number;
            /** Format: int64 */
            total: number;
        };
        UpdateStatusInputBody: {
            /** @enum {string} */
            status: "new" | "learning" | "learned";
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaAddUnitInputBody = components['schemas']['AddUnitInputBody'];
export type SchemaAddUnitOutput = components['schemas']['AddUnitOutput'];
export type SchemaAnalyzeInputBody = components['schemas']['AnalyzeInputBody'];
export type SchemaAnalyzeItemOutput = components['schemas']['AnalyzeItemOutput'];
export type SchemaAnalyzeOutput = components['schemas']['AnalyzeOutput'];
export type SchemaAnswerCardInputBody = components['schemas']['AnswerCardInputBody'];
export type SchemaAnswerCardOutput = components['schemas']['AnswerCardOutput'];
export type SchemaCardOutput = components['schemas']['CardOutput'];
export type SchemaCardsOutput = components['schemas']['CardsOutput'];
export type SchemaCheckPracticeInputBody = components['schemas']['CheckPracticeInputBody'];
export type SchemaCheckPracticeOutput = components['schemas']['CheckPracticeOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaDeleteUnitOutput = components['schemas']['DeleteUnitOutput'];
export type SchemaDueCardsOutput = components['schemas']['DueCardsOutput'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaGenerateCardsInputBody = components['schemas']['GenerateCardsInputBody'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaOccurrenceOutput = components['schemas']['OccurrenceOutput'];
export type SchemaOccurrencesOutput = components['schemas']['OccurrencesOutput'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaPracticeAnswerDto = components['schemas']['PracticeAnswerDTO'];
export type SchemaPracticeOutput = components['schemas']['PracticeOutput'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSavePracticeProgressInputBody = components['schemas']['SavePracticeProgressInputBody'];
export type SchemaSavePracticeProgressOutput = components['schemas']['SavePracticeProgressOutput'];
export type SchemaSpotOutput = components['schemas']['SpotOutput'];
export type SchemaSuccessBodyAddUnitOutput = components['schemas']['SuccessBodyAddUnitOutput'];
export type SchemaSuccessBodyAnalyzeOutput = components['schemas']['SuccessBodyAnalyzeOutput'];
export type SchemaSuccessBodyAnswerCardOutput = components['schemas']['SuccessBodyAnswerCardOutput'];
export type SchemaSuccessBodyCardsOutput = components['schemas']['SuccessBodyCardsOutput'];
export type SchemaSuccessBodyCheckPracticeOutput = components['schemas']['SuccessBodyCheckPracticeOutput'];
export type SchemaSuccessBodyDeleteUnitOutput = components['schemas']['SuccessBodyDeleteUnitOutput'];
export type SchemaSuccessBodyDueCardsOutput = components['schemas']['SuccessBodyDueCardsOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyOccurrencesOutput = components['schemas']['SuccessBodyOccurrencesOutput'];
export type SchemaSuccessBodyPracticeOutput = components['schemas']['SuccessBodyPracticeOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodySavePracticeProgressOutput = components['schemas']['SuccessBodySavePracticeProgressOutput'];
export type SchemaSuccessBodyUnitOutput = components['schemas']['SuccessBodyUnitOutput'];
export type SchemaSuccessBodyUnitsOutput = components['schemas']['SuccessBodyUnitsOutput'];
export type SchemaUnitOutput = components['schemas']['UnitOutput'];
export type SchemaUnitsOutput = components['schemas']['UnitsOutput'];
export type SchemaUpdateStatusInputBody = components['schemas']['UpdateStatusInputBody'];
export type $defs = Record<string, never>;
export interface operations {
    analyzeVocab: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["AnalyzeInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyAnalyzeOutput"];
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
    generateVocabCards: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["GenerateCardsInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyCardsOutput"];
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
    answerVocabCard: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["AnswerCardInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyAnswerCardOutput"];
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
    dueVocabCards: {
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
                    "application/json": components["schemas"]["SuccessBodyDueCardsOutput"];
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
    vocabOccurrences: {
        parameters: {
            query: {
                /** @description Word or phrase to look up in the parsed lexicon */
                text: string;
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
                    "application/json": components["schemas"]["SuccessBodyOccurrencesOutput"];
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
    getVocabPractice: {
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
                    "application/json": components["schemas"]["SuccessBodyPracticeOutput"];
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
    generateVocabPractice: {
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
                    "application/json": components["schemas"]["SuccessBodyPracticeOutput"];
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
    checkVocabPractice: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CheckPracticeInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyCheckPracticeOutput"];
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
    saveVocabPracticeProgress: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SavePracticeProgressInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodySavePracticeProgressOutput"];
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
    listVocabUnits: {
        parameters: {
            query?: {
                q?: string;
                status?: "" | "new" | "learning" | "learned";
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
    addVocabUnit: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["AddUnitInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyAddUnitOutput"];
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
    deleteVocabUnit: {
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
    updateVocabUnitStatus: {
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
                "application/json": components["schemas"]["UpdateStatusInputBody"];
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
}
