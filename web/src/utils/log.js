
var DODEBUG;

/// Load configuration if it is possible to do so.
if( (typeof window!=='undefined') && window.____API_WEB_CONFIG____ ) {
  DODEBUG = !!window.____API_WEB_CONFIG____.DEBUG;
} else {
  DODEBUG = true;
}

let Logger = function(DODEBUG) {

  return {
    /// Parse to preserve the given object status for logging.
    o: function(o) {
      return Object.assign({}, o);
    },
    withPrefix: function(prefix) {

      var O = {};

      Object.defineProperty(O, "lin", {
          get: function() {
              return function(mid) { return Logger(DODEBUG).withPrefix(prefix+mid); };
          },
          configurable: false,
          enumerable: false
      });

      Object.defineProperty(O, "makeTimer", {
          get: function() {
              return function(uid) {
                uid = uid || (new Date).getTime();
                return O.lin(`[timer][${uid}]`);
              };
          },
          configurable: false,
          enumerable: false
      });

      Object.defineProperty(O, "o", {
          get: function() {
              return Logger.o;
          },
          configurable: false,
          enumerable: false
      });

      // Disable all logging when not in debug mode.
      if(!DODEBUG) {

        for (var m in console) {
          let fn = console[m];
          if (typeof fn == 'function') {

            Object.defineProperty(O, m, {
                get: function() {
                    return ()=>{};
                },
                configurable: false,
                enumerable: false
            });

          }
        }
        return O;

      }


      const mixinFns = [
        "info",
        "log",
        "debug",
        "warn",
        "error",
        // Timing function will be called with prefix as first argument, thus the naming.
        "time",
        "timeLog",
        "timeEnd",
        // Grouping functions
        "group",
        "groupCollapsed",
        "groupEnd"
      ];

      for (var m2 in console) {
        let fn = console[m2];
        if (typeof fn == 'function') {

          if( mixinFns.indexOf(m2) >= 0 ) {

            Object.defineProperty(O, m2, {
                get: function() {
                    return fn.bind(console, prefix);
                },
                configurable: false,
                enumerable: false
            });

          } else {

            // Preserve other functions.
            Object.defineProperty(O, m2, {
                get: function() {
                    return console[m2].bind(console);
                },
                configurable: false,
                enumerable: false
            });

          }

        }
      }


      return O;

    }
  };

};

export default Logger(DODEBUG);

export {
  Logger
}