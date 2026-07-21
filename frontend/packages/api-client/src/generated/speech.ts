export interface paths {
    "/api/v1/speech/assess": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Score the pronunciation of a recorded reading against the reference text */
        post: operations["assessSpeech"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/speech/feedback": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Get LLM coaching advice for a scored reading */
        post: operations["speechFeedback"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/speech/phonemes": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List the phoneme articulation guide */
        get: operations["listSpeechPhonemes"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/speech/practice": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Generate sentences to practice reading aloud */
        post: operations["speechGeneratePractice"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/speech/tts": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Synthesize speech audio for a text */
        post: operations["speechSynthesize"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/speech/voices": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List available TTS voices */
        get: operations["listSpeechVoices"];
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
        AssessOutput: {
            heard: string;
            /** Format: int64 */
            overall: number;
            words: components["schemas"]["WordOutput"][] | null;
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
        FeedbackInputBody: {
            /** @description IPA transcription of what was actually said */
            heard: string;
            /** @description Detected per-phoneme issues, e.g. "think: expected θ, heard s" */
            issues?: string[] | null;
            /** @description Student's native language for advice wording */
            native_language: string;
            /** @description Reference text that was read aloud */
            text: string;
        };
        FeedbackOutput: {
            summary: string;
            tips: components["schemas"]["FeedbackTipOutput"][] | null;
        };
        FeedbackTipOutput: {
            advice: string;
            sound: string;
        };
        FormFile: {
            ContentType: string;
            Filename: string;
            IsSet: boolean;
            /** Format: int64 */
            Size: number;
        };
        GeneratePracticeInputBody: {
            /** @description Target IPA sounds to pack into the sentences */
            sounds?: string[] | null;
            /** @description Optional topic for the sentences */
            topic?: string;
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
        PhonemeInfoOutput: {
            description: string;
            examples: string;
            /** @enum {string} */
            kind: "vowel" | "diphthong" | "consonant";
            pitfall?: string;
            symbol: string;
        };
        PhonemeOutput: {
            expected: string;
            heard?: string;
            /** Format: double */
            score: number;
            /** @enum {string} */
            verdict: "good" | "close" | "wrong" | "missing";
        };
        PhonemesOutput: {
            items: components["schemas"]["PhonemeInfoOutput"][] | null;
        };
        ReadyOutput: {
            checks?: components["schemas"]["CheckResult"][] | null;
            /** @example auth */
            module?: string;
            ready: boolean;
            /** Format: date-time */
            time: string;
        };
        SpeechPracticeOutput: {
            sentences: string[] | null;
        };
        SuccessBodyAssessOutput: {
            data: components["schemas"]["AssessOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyFeedbackOutput: {
            data: components["schemas"]["FeedbackOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyPhonemesOutput: {
            data: components["schemas"]["PhonemesOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodySpeechPracticeOutput: {
            data: components["schemas"]["SpeechPracticeOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodySynthesizeOutput: {
            data: components["schemas"]["SynthesizeOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyVoicesOutput: {
            data: components["schemas"]["VoicesOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SynthesizeInputBody: {
            /**
             * Format: double
             * @description Speech speed multiplier, default 1.0
             */
            speed?: number;
            /** @description Text to speak aloud */
            text: string;
            /** @description Voice name; a random one is used when omitted */
            voice?: string;
        };
        SynthesizeOutput: {
            /** @description WAV audio, base64-encoded */
            audio_base64: string;
            /** @description Voice that was actually used */
            voice: string;
        };
        VoicesOutput: {
            voices: string[] | null;
        };
        WordOutput: {
            extra: string[] | null;
            ipa: string;
            phonemes: components["schemas"]["PhonemeOutput"][] | null;
            /** Format: int64 */
            score: number;
            word: string;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaAssessOutput = components['schemas']['AssessOutput'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaFeedbackInputBody = components['schemas']['FeedbackInputBody'];
export type SchemaFeedbackOutput = components['schemas']['FeedbackOutput'];
export type SchemaFeedbackTipOutput = components['schemas']['FeedbackTipOutput'];
export type SchemaFormFile = components['schemas']['FormFile'];
export type SchemaGeneratePracticeInputBody = components['schemas']['GeneratePracticeInputBody'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaPhonemeInfoOutput = components['schemas']['PhonemeInfoOutput'];
export type SchemaPhonemeOutput = components['schemas']['PhonemeOutput'];
export type SchemaPhonemesOutput = components['schemas']['PhonemesOutput'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaSpeechPracticeOutput = components['schemas']['SpeechPracticeOutput'];
export type SchemaSuccessBodyAssessOutput = components['schemas']['SuccessBodyAssessOutput'];
export type SchemaSuccessBodyFeedbackOutput = components['schemas']['SuccessBodyFeedbackOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyPhonemesOutput = components['schemas']['SuccessBodyPhonemesOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodySpeechPracticeOutput = components['schemas']['SuccessBodySpeechPracticeOutput'];
export type SchemaSuccessBodySynthesizeOutput = components['schemas']['SuccessBodySynthesizeOutput'];
export type SchemaSuccessBodyVoicesOutput = components['schemas']['SuccessBodyVoicesOutput'];
export type SchemaSynthesizeInputBody = components['schemas']['SynthesizeInputBody'];
export type SchemaSynthesizeOutput = components['schemas']['SynthesizeOutput'];
export type SchemaVoicesOutput = components['schemas']['VoicesOutput'];
export type SchemaWordOutput = components['schemas']['WordOutput'];
export type $defs = Record<string, never>;
export interface operations {
    assessSpeech: {
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
                    audio: string;
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
                    "application/json": components["schemas"]["SuccessBodyAssessOutput"];
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
    speechFeedback: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["FeedbackInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyFeedbackOutput"];
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
    listSpeechPhonemes: {
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
                    "application/json": components["schemas"]["SuccessBodyPhonemesOutput"];
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
    speechGeneratePractice: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["GeneratePracticeInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodySpeechPracticeOutput"];
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
    speechSynthesize: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SynthesizeInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodySynthesizeOutput"];
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
    listSpeechVoices: {
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
                    "application/json": components["schemas"]["SuccessBodyVoicesOutput"];
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
