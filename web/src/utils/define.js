
/**
 * Basic tools for Model Defining Code
 */

const DefineStatic = function(obj, name, val) {
    Object.defineProperty(obj, name, {
        get: ()=>val
    });
}

const ModelProperty = function(obj, name, opts) {
    Object.defineProperty(obj, name, {
        ...opts,
        configurable: true,
        enumerable: true
    });
};


const DEFINE = function(obj) {

    let DEF = {};
    let wrapper = function(fn) {
        return function() {
            fn.apply(undefined, arguments);
            return DEF;
        };
    }

    DefineStatic(DEF, "static", wrapper((name,val)=>DefineStatic(obj,name,val)));
    DefineStatic(DEF, "property", wrapper((name,opts)=>ModelProperty(obj,name,opts)));

    return DEF;

}

export default DEFINE;