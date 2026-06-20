export namespace ippool {
	
	export class MobileConflictItem {
	    reason: string;
	    iosEvidenceIds?: string[];
	    androidEvidenceIds?: string[];
	    severity: string;
	    resolutionHint: string;
	
	    static createFrom(source: any = {}) {
	        return new MobileConflictItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reason = source["reason"];
	        this.iosEvidenceIds = source["iosEvidenceIds"];
	        this.androidEvidenceIds = source["androidEvidenceIds"];
	        this.severity = source["severity"];
	        this.resolutionHint = source["resolutionHint"];
	    }
	}
	export class MobileEvidenceItem {
	    id: string;
	    type: string;
	    value: string;
	    osHint: string;
	    confidenceImpact: number;
	    strength: string;
	    source: string;
	    timestamp: string;
	    explanation: string;
	
	    static createFrom(source: any = {}) {
	        return new MobileEvidenceItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.value = source["value"];
	        this.osHint = source["osHint"];
	        this.confidenceImpact = source["confidenceImpact"];
	        this.strength = source["strength"];
	        this.source = source["source"];
	        this.timestamp = source["timestamp"];
	        this.explanation = source["explanation"];
	    }
	}
	export class MobileFingerprint {
	    classification: string;
	    iosScore: number;
	    androidScore: number;
	    ipadScore: number;
	    confidence: number;
	    evidence?: MobileEvidenceItem[];
	    conflicts?: MobileConflictItem[];
	    warnings?: string[];
	    lastUpdatedAt?: string;
	    whyThisClassification?: string;
	    whyNotCertain?: string;
	
	    static createFrom(source: any = {}) {
	        return new MobileFingerprint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.classification = source["classification"];
	        this.iosScore = source["iosScore"];
	        this.androidScore = source["androidScore"];
	        this.ipadScore = source["ipadScore"];
	        this.confidence = source["confidence"];
	        this.evidence = this.convertValues(source["evidence"], MobileEvidenceItem);
	        this.conflicts = this.convertValues(source["conflicts"], MobileConflictItem);
	        this.warnings = source["warnings"];
	        this.lastUpdatedAt = source["lastUpdatedAt"];
	        this.whyThisClassification = source["whyThisClassification"];
	        this.whyNotCertain = source["whyNotCertain"];
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
	export class EvidenceItem {
	    type: string;
	    source: string;
	    value: string;
	    timestamp: string;
	    confidenceImpact: number;
	    strength: string;
	
	    static createFrom(source: any = {}) {
	        return new EvidenceItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.source = source["source"];
	        this.value = source["value"];
	        this.timestamp = source["timestamp"];
	        this.confidenceImpact = source["confidenceImpact"];
	        this.strength = source["strength"];
	    }
	}
	export class DevicePoolEntry {
	    id: string;
	    ip: string;
	    mac?: string;
	    hostname?: string;
	    vendor?: string;
	    firstSeen: string;
	    lastSeen?: string;
	    lastProbeAt?: string;
	    status: string;
	    responseCount: number;
	    failureCount: number;
	    avgLatencyMs?: number;
	    ttl?: number;
	    source: string;
	    evidence?: EvidenceItem[];
	    mobileFingerprint?: MobileFingerprint;
	    deviceTypeHint?: string;
	    osHint?: string;
	    osConfidence?: number;
	    osEvidenceSummary?: string[];
	
	    static createFrom(source: any = {}) {
	        return new DevicePoolEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ip = source["ip"];
	        this.mac = source["mac"];
	        this.hostname = source["hostname"];
	        this.vendor = source["vendor"];
	        this.firstSeen = source["firstSeen"];
	        this.lastSeen = source["lastSeen"];
	        this.lastProbeAt = source["lastProbeAt"];
	        this.status = source["status"];
	        this.responseCount = source["responseCount"];
	        this.failureCount = source["failureCount"];
	        this.avgLatencyMs = source["avgLatencyMs"];
	        this.ttl = source["ttl"];
	        this.source = source["source"];
	        this.evidence = this.convertValues(source["evidence"], EvidenceItem);
	        this.mobileFingerprint = this.convertValues(source["mobileFingerprint"], MobileFingerprint);
	        this.deviceTypeHint = source["deviceTypeHint"];
	        this.osHint = source["osHint"];
	        this.osConfidence = source["osConfidence"];
	        this.osEvidenceSummary = source["osEvidenceSummary"];
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
	
	
	

}

