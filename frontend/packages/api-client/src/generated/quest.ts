export interface paths {
    "/api/v1/quest/missions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List quest missions with generation status */
        get: operations["listQuestMissions"];
        put?: never;
        /** Create a quest mission */
        post: operations["createQuestMission"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/quest/missions/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get a quest mission (polling: generation, images, active reply) */
        get: operations["getQuestMission"];
        put?: never;
        post?: never;
        /** Delete a quest mission */
        delete: operations["deleteQuestMission"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/quest/missions/{id}/native-reply": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Suggest native-like variants for the player reply */
        post: operations["suggestQuestNativeReply"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/quest/missions/{id}/regenerate-images": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Regenerate failed mission images (cover, scenes, avatars) */
        post: operations["regenerateQuestMissionImages"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/quest/missions/{id}/reset": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Reset a quest mission to its first scene */
        post: operations["resetQuestMission"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/quest/missions/{id}/respond": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Submit a player response (async) */
        post: operations["respondQuestMission"];
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
        Character: {
            age?: string;
            appearance: string;
            arc: string;
            gender?: string;
            /** Format: int64 */
            initialTrust?: number;
            motivation: string;
            name: string;
            personality: string;
            role: string;
            speechStyle: string;
            voice?: string;
        };
        CharacterLine: {
            name: string;
            text: string;
            voice?: string;
        };
        CheckResult: {
            error?: string;
            /** @example postgres */
            name: string;
            ok: boolean;
        };
        CreateMissionInputBody: {
            genre?: string;
            language?: string;
            practiceGoals?: string;
            prompt?: string;
        };
        CreateMissionOutput: {
            missionId: string;
        };
        CustomMission: {
            characterAvatarErrors?: {
                [key: string]: string;
            };
            characterAvatarGenStartedAt?: {
                [key: string]: string;
            };
            characterAvatarStatus?: {
                [key: string]: string;
            };
            characterAvatars?: {
                [key: string]: string;
            };
            characters: components["schemas"]["Character"][] | null;
            coverImage?: string;
            coverImageError?: string;
            coverImageGenStartedAt?: string;
            coverImageStatus?: string;
            createdAt: string;
            currentScene: components["schemas"]["DynamicScene"];
            /** Format: int64 */
            currentStage: number;
            description: string;
            /** Format: int64 */
            estimatedScenes?: number;
            generationError?: string;
            generationStatus?: string;
            generationStep?: string;
            genre: string;
            history: components["schemas"]["DialogueTurn"][] | null;
            historySummary?: string;
            id: string;
            isComplete: boolean;
            language: string;
            narratorVoice?: string;
            npcStates?: {
                [key: string]: components["schemas"]["NPCState"];
            };
            outcome?: string;
            playerAvatarImage?: string;
            plotPoints?: components["schemas"]["PlotPoint"][] | null;
            practiceGoals: string;
            resolution?: components["schemas"]["Resolution"];
            sceneImageErrors?: {
                [key: string]: string;
            };
            sceneImageGenStartedAt?: {
                [key: string]: string;
            };
            sceneImageStatus?: {
                [key: string]: string;
            };
            sceneImages?: {
                [key: string]: string;
            };
            scenes: components["schemas"]["DynamicScene"][] | null;
            secretEnding: string;
            skillCategories?: {
                [key: string]: string;
            };
            skillSignals?: {
                [key: string]: number;
            };
            skillsEarned: components["schemas"]["SkillReward"][] | null;
            /** Format: int64 */
            summarizedUpToTurn?: number;
            title: string;
            /** Format: int64 */
            totalStages: number;
            /** Format: int64 */
            totalXp?: number;
            userPrompt: string;
        };
        DeleteMissionOutput: {
            ok: boolean;
        };
        DialogueTurn: {
            /** Format: int64 */
            scene: number;
            speaker: string;
            text: string;
            voice?: string;
        };
        DynamicScene: {
            flavor?: string;
            isFinal: boolean;
            narration: string;
            narrationVoice?: string;
            objects: string[] | null;
            present: components["schemas"]["SceneCharacter"][] | null;
            scenePurpose?: string;
            /** Format: int64 */
            stage: number;
            summary?: string;
            tips: components["schemas"]["LanguageTip"][] | null;
            trigger?: string;
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
        GetMissionOutput: {
            activeReply?: components["schemas"]["RespondJobStatusResponse"];
            mission: components["schemas"]["CustomMission"];
        };
        GrammarCheck: {
            errors: components["schemas"]["GrammarError"][] | null;
            ok: boolean;
        };
        GrammarError: {
            correction: string;
            explanation: string;
            original: string;
            type: string;
        };
        HealthOutput: {
            /** @example auth */
            module?: string;
            /** @example ok */
            status: string;
            /** Format: date-time */
            time: string;
        };
        LanguageTip: {
            category: string;
            construction: string;
            example: string;
            explanation: string;
            tip: string;
        };
        ListMissionsOutput: {
            missions: components["schemas"]["MissionSummary"][] | null;
        };
        Meta: {
            pagination?: components["schemas"]["Pagination"];
            request_id?: string;
        };
        MissionSummary: {
            coverImage?: string;
            coverImageStatus?: string;
            createdAt: string;
            /** Format: int64 */
            currentStage: number;
            description: string;
            generationError?: string;
            generationStatus: string;
            generationStep?: string;
            genre?: string;
            id: string;
            isComplete: boolean;
            language?: string;
            started: boolean;
            title: string;
            /** Format: int64 */
            totalStages: number;
        };
        NPCState: {
            knowsAboutPlayer?: string[] | null;
            playerKnowsAbout?: string[] | null;
            /** Format: int64 */
            trust: number;
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
        PartialLine: {
            done?: boolean;
            name: string;
            text?: string;
        };
        PartialWorld: {
            narration?: string;
            narrationDone?: boolean;
            responses?: components["schemas"]["PartialLine"][] | null;
        };
        PlotPoint: {
            delivered: boolean;
            /** Format: int64 */
            deliveredAt?: number;
            description?: string;
            fact: string;
            id: string;
            required: boolean;
        };
        ReadyOutput: {
            checks?: components["schemas"]["CheckResult"][] | null;
            /** @example auth */
            module?: string;
            ready: boolean;
            /** Format: date-time */
            time: string;
        };
        RegenerateImagesInputBody: {
            key?: string;
            /** @enum {string} */
            kind?: "" | "cover" | "scene" | "avatar";
        };
        RegenerateImagesOutput: {
            mission: components["schemas"]["CustomMission"];
        };
        ResetMissionOutput: {
            mission: components["schemas"]["CustomMission"];
        };
        Resolution: {
            goal: string;
            outcomes: components["schemas"]["ResolutionOutcome"][] | null;
            type: string;
        };
        ResolutionOutcome: {
            description: string;
            label: string;
        };
        RespondInputBody: {
            strict?: boolean;
            text: string;
        };
        RespondJobResult: {
            /** Format: int64 */
            currentStage: number;
            epilogue?: string;
            errors: components["schemas"]["GrammarError"][] | null;
            grammarOk: boolean;
            isComplete: boolean;
            narration: string;
            narrationVoice?: string;
            nextScene?: components["schemas"]["DynamicScene"];
            outcome?: string;
            playerIntent?: string;
            responses: components["schemas"]["CharacterLine"][] | null;
            sceneAdvanced: boolean;
            sceneState?: string;
            /** Format: int64 */
            totalStages: number;
        };
        RespondJobStatusResponse: {
            error?: string;
            grammar?: components["schemas"]["GrammarCheck"];
            inputText?: string;
            jobId: string;
            partial?: components["schemas"]["PartialWorld"];
            result?: components["schemas"]["RespondJobResult"];
            status: string;
            step: string;
        };
        RespondOutput: {
            jobId: string;
        };
        SceneCharacter: {
            dialogue: string;
            name: string;
            voice?: string;
        };
        SkillReward: {
            category: string;
            isNew: boolean;
            name: string;
        };
        SuccessBodyCreateMissionOutput: {
            data: components["schemas"]["CreateMissionOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyDeleteMissionOutput: {
            data: components["schemas"]["DeleteMissionOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyGetMissionOutput: {
            data: components["schemas"]["GetMissionOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyHealthOutput: {
            data: components["schemas"]["HealthOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyListMissionsOutput: {
            data: components["schemas"]["ListMissionsOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyReadyOutput: {
            data: components["schemas"]["ReadyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyRegenerateImagesOutput: {
            data: components["schemas"]["RegenerateImagesOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyResetMissionOutput: {
            data: components["schemas"]["ResetMissionOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodyRespondOutput: {
            data: components["schemas"]["RespondOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuccessBodySuggestNativeReplyOutput: {
            data: components["schemas"]["SuggestNativeReplyOutput"];
            meta?: components["schemas"]["Meta"];
            ok: boolean;
        };
        SuggestNativeReplyInputBody: {
            text: string;
        };
        SuggestNativeReplyOutput: {
            variants: string[] | null;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaCharacter = components['schemas']['Character'];
export type SchemaCharacterLine = components['schemas']['CharacterLine'];
export type SchemaCheckResult = components['schemas']['CheckResult'];
export type SchemaCreateMissionInputBody = components['schemas']['CreateMissionInputBody'];
export type SchemaCreateMissionOutput = components['schemas']['CreateMissionOutput'];
export type SchemaCustomMission = components['schemas']['CustomMission'];
export type SchemaDeleteMissionOutput = components['schemas']['DeleteMissionOutput'];
export type SchemaDialogueTurn = components['schemas']['DialogueTurn'];
export type SchemaDynamicScene = components['schemas']['DynamicScene'];
export type SchemaErrorBody = components['schemas']['ErrorBody'];
export type SchemaErrorDetail = components['schemas']['ErrorDetail'];
export type SchemaErrorResponse = components['schemas']['ErrorResponse'];
export type SchemaGetMissionOutput = components['schemas']['GetMissionOutput'];
export type SchemaGrammarCheck = components['schemas']['GrammarCheck'];
export type SchemaGrammarError = components['schemas']['GrammarError'];
export type SchemaHealthOutput = components['schemas']['HealthOutput'];
export type SchemaLanguageTip = components['schemas']['LanguageTip'];
export type SchemaListMissionsOutput = components['schemas']['ListMissionsOutput'];
export type SchemaMeta = components['schemas']['Meta'];
export type SchemaMissionSummary = components['schemas']['MissionSummary'];
export type SchemaNpcState = components['schemas']['NPCState'];
export type SchemaPagination = components['schemas']['Pagination'];
export type SchemaPartialLine = components['schemas']['PartialLine'];
export type SchemaPartialWorld = components['schemas']['PartialWorld'];
export type SchemaPlotPoint = components['schemas']['PlotPoint'];
export type SchemaReadyOutput = components['schemas']['ReadyOutput'];
export type SchemaRegenerateImagesInputBody = components['schemas']['RegenerateImagesInputBody'];
export type SchemaRegenerateImagesOutput = components['schemas']['RegenerateImagesOutput'];
export type SchemaResetMissionOutput = components['schemas']['ResetMissionOutput'];
export type SchemaResolution = components['schemas']['Resolution'];
export type SchemaResolutionOutcome = components['schemas']['ResolutionOutcome'];
export type SchemaRespondInputBody = components['schemas']['RespondInputBody'];
export type SchemaRespondJobResult = components['schemas']['RespondJobResult'];
export type SchemaRespondJobStatusResponse = components['schemas']['RespondJobStatusResponse'];
export type SchemaRespondOutput = components['schemas']['RespondOutput'];
export type SchemaSceneCharacter = components['schemas']['SceneCharacter'];
export type SchemaSkillReward = components['schemas']['SkillReward'];
export type SchemaSuccessBodyCreateMissionOutput = components['schemas']['SuccessBodyCreateMissionOutput'];
export type SchemaSuccessBodyDeleteMissionOutput = components['schemas']['SuccessBodyDeleteMissionOutput'];
export type SchemaSuccessBodyGetMissionOutput = components['schemas']['SuccessBodyGetMissionOutput'];
export type SchemaSuccessBodyHealthOutput = components['schemas']['SuccessBodyHealthOutput'];
export type SchemaSuccessBodyListMissionsOutput = components['schemas']['SuccessBodyListMissionsOutput'];
export type SchemaSuccessBodyReadyOutput = components['schemas']['SuccessBodyReadyOutput'];
export type SchemaSuccessBodyRegenerateImagesOutput = components['schemas']['SuccessBodyRegenerateImagesOutput'];
export type SchemaSuccessBodyResetMissionOutput = components['schemas']['SuccessBodyResetMissionOutput'];
export type SchemaSuccessBodyRespondOutput = components['schemas']['SuccessBodyRespondOutput'];
export type SchemaSuccessBodySuggestNativeReplyOutput = components['schemas']['SuccessBodySuggestNativeReplyOutput'];
export type SchemaSuggestNativeReplyInputBody = components['schemas']['SuggestNativeReplyInputBody'];
export type SchemaSuggestNativeReplyOutput = components['schemas']['SuggestNativeReplyOutput'];
export type $defs = Record<string, never>;
export interface operations {
    listQuestMissions: {
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
                    "application/json": components["schemas"]["SuccessBodyListMissionsOutput"];
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
    createQuestMission: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CreateMissionInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyCreateMissionOutput"];
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
    getQuestMission: {
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
                    "application/json": components["schemas"]["SuccessBodyGetMissionOutput"];
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
    deleteQuestMission: {
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
                    "application/json": components["schemas"]["SuccessBodyDeleteMissionOutput"];
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
    suggestQuestNativeReply: {
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
                "application/json": components["schemas"]["SuggestNativeReplyInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodySuggestNativeReplyOutput"];
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
    regenerateQuestMissionImages: {
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
                "application/json": components["schemas"]["RegenerateImagesInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyRegenerateImagesOutput"];
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
    resetQuestMission: {
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
                    "application/json": components["schemas"]["SuccessBodyResetMissionOutput"];
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
    respondQuestMission: {
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
                "application/json": components["schemas"]["RespondInputBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SuccessBodyRespondOutput"];
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
