export namespace main {
	
	export class CacheInfo {
	    size: number;
	    sizeHuman: string;
	
	    static createFrom(source: any = {}) {
	        return new CacheInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.size = source["size"];
	        this.sizeHuman = source["sizeHuman"];
	    }
	}
	export class DroppedFile {
	    name: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new DroppedFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.content = source["content"];
	    }
	}

}

export namespace models {
	
	export class ProviderConfig {
	    apiKey: string;
	    model: string;
	    models: string[];
	    defaultModel: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.models = source["models"];
	        this.defaultModel = source["defaultModel"];
	        this.enabled = source["enabled"];
	    }
	}
	export class AppConfig {
	    providers: Record<string, ProviderConfig>;
	    activeProvider: string;
	    workerCount: number;
	    chunkSize: number;
	    enableReview: boolean;
	    outputDir: string;
	    outputFormat: string;
	    fallbackOnFailure: boolean;
	    theme: string;
	    customPromptDir: string;
	    customPromptDirEnabled: boolean;
	    maxRetries: number;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providers = this.convertValues(source["providers"], ProviderConfig, true);
	        this.activeProvider = source["activeProvider"];
	        this.workerCount = source["workerCount"];
	        this.chunkSize = source["chunkSize"];
	        this.enableReview = source["enableReview"];
	        this.outputDir = source["outputDir"];
	        this.outputFormat = source["outputFormat"];
	        this.fallbackOnFailure = source["fallbackOnFailure"];
	        this.theme = source["theme"];
	        this.customPromptDir = source["customPromptDir"];
	        this.customPromptDirEnabled = source["customPromptDirEnabled"];
	        this.maxRetries = source["maxRetries"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    size: number;
	    isDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.isDir = source["isDir"];
	    }
	}
	export class GenrePrompt {
	    id: string;
	    name: string;
	    content: string;
	    builtin: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GenrePrompt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.content = source["content"];
	        this.builtin = source["builtin"];
	    }
	}
	export class GlossaryEntry {
	    original: string;
	    translated: string;
	    category: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new GlossaryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.original = source["original"];
	        this.translated = source["translated"];
	        this.category = source["category"];
	        this.note = source["note"];
	    }
	}
	export class Glossary {
	    novelName: string;
	    author: string;
	    genre: string;
	    entries: Record<string, GlossaryEntry>;
	
	    static createFrom(source: any = {}) {
	        return new Glossary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.novelName = source["novelName"];
	        this.author = source["author"];
	        this.genre = source["genre"];
	        this.entries = this.convertValues(source["entries"], GlossaryEntry, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class TestConnectionResult {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new TestConnectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class TranslationTask {
	    id: string;
	    name: string;
	    genre: string;
	    inputPath: string;
	    outputPath: string;
	    status: string;
	    totalChunks: number;
	    doneChunks: number;
	    currentChunk: number;
	    progress: number;
	    error: string;
	    startedAt: string;
	    finishedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TranslationTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.genre = source["genre"];
	        this.inputPath = source["inputPath"];
	        this.outputPath = source["outputPath"];
	        this.status = source["status"];
	        this.totalChunks = source["totalChunks"];
	        this.doneChunks = source["doneChunks"];
	        this.currentChunk = source["currentChunk"];
	        this.progress = source["progress"];
	        this.error = source["error"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	    }
	}

}

