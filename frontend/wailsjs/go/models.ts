export namespace main {
	
	export class TestParams {
	    pkg: string;
	    verbose: boolean;
	    race: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TestParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pkg = source["pkg"];
	        this.verbose = source["verbose"];
	        this.race = source["race"];
	    }
	}

}

