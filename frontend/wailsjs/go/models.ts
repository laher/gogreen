export namespace main {
	
	export class Package {
	    pkg: string;
	    testFuncs: string[];
	
	    static createFrom(source: any = {}) {
	        return new Package(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pkg = source["pkg"];
	        this.testFuncs = source["testFuncs"];
	    }
	}
	export class TestParams {
	    pkg: string;
	    verbose: boolean;
	    race: boolean;
	    run: string;
	
	    static createFrom(source: any = {}) {
	        return new TestParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pkg = source["pkg"];
	        this.verbose = source["verbose"];
	        this.race = source["race"];
	        this.run = source["run"];
	    }
	}
	export class State {
	    cwd: string;
	    pkg_list: string[];
	    watching: boolean;
	    running: boolean;
	    test_params: TestParams;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cwd = source["cwd"];
	        this.pkg_list = source["pkg_list"];
	        this.watching = source["watching"];
	        this.running = source["running"];
	        this.test_params = this.convertValues(source["test_params"], TestParams);
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

