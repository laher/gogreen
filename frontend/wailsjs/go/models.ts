export namespace main {
	
	export class State {
	    cwd: string;
	    pkg_list: string[];
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cwd = source["cwd"];
	        this.pkg_list = source["pkg_list"];
	    }
	}
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

