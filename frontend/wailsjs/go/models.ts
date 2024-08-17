export namespace main {
	
	export class TestParams {
	    pkg: string;
	
	    static createFrom(source: any = {}) {
	        return new TestParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pkg = source["pkg"];
	    }
	}

}

