// Reads and forms basic config for WebUI

import DEFINE from "./utils/define"
import log from "./utils/log"

const logger = log.withPrefix("[config]");

const Host = function(port) {
    return "//" + window.location.hostname + ":" + port + "";
}

const CONFIG = function() {

    if(window && window.____API_WEB_CONFIG____) {

        const C = {}

        const host = Host()

        let port = ____API_WEB_CONFIG____.apiPort

        DEFINE(C)
        .static("apiPort", port)
        .static("hostUrl", Host(port))

        return C;

    } else {
        logger.error("Failed to load Web UI config.")
        return {}
    }

}

export default CONFIG()