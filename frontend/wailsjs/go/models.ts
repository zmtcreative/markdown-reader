export namespace app {
	
	export class AlertCalloutOptions {
	    use_alertcallouts: boolean;
	    alertcallout_style: string;
	
	    static createFrom(source: any = {}) {
	        return new AlertCalloutOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.use_alertcallouts = source["use_alertcallouts"];
	        this.alertcallout_style = source["alertcallout_style"];
	    }
	}
	export class ApplicationOptions {
	    use_inline_html: boolean;
	    use_sanitize_html: boolean;
	    strip_h1: boolean;
	    use_frontmatter: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ApplicationOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.use_inline_html = source["use_inline_html"];
	        this.use_sanitize_html = source["use_sanitize_html"];
	        this.strip_h1 = source["strip_h1"];
	        this.use_frontmatter = source["use_frontmatter"];
	    }
	}
	export class MarkdownOptions {
	    use_gfm: boolean;
	    use_emoji: boolean;
	    use_mermaid: boolean;
	    use_figure: boolean;
	    use_anchor: boolean;
	    use_fences: boolean;
	    use_sections: boolean;
	    use_highlighting: boolean;
	    use_fancylists: boolean;
	    use_attributes: boolean;
	    use_typographic: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MarkdownOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.use_gfm = source["use_gfm"];
	        this.use_emoji = source["use_emoji"];
	        this.use_mermaid = source["use_mermaid"];
	        this.use_figure = source["use_figure"];
	        this.use_anchor = source["use_anchor"];
	        this.use_fences = source["use_fences"];
	        this.use_sections = source["use_sections"];
	        this.use_highlighting = source["use_highlighting"];
	        this.use_fancylists = source["use_fancylists"];
	        this.use_attributes = source["use_attributes"];
	        this.use_typographic = source["use_typographic"];
	    }
	}
	export class Config {
	    application: ApplicationOptions;
	    markdown: MarkdownOptions;
	    alert_callouts: AlertCalloutOptions;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.application = this.convertValues(source["application"], ApplicationOptions);
	        this.markdown = this.convertValues(source["markdown"], MarkdownOptions);
	        this.alert_callouts = this.convertValues(source["alert_callouts"], AlertCalloutOptions);
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

