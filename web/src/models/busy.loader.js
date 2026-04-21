
import { ref } from "vue"
import DEFINE from "@/utils/define"

const BusyLoader = function(){

    const O = {};

    let loading = ref(false);
    let fading;

    const busy = function() {
        loading.value = true;
        fading = false;
    }

    const idle = function() {
        fading = true;
        setTimeout(()=>{
            if(fading) {
                loading.value = false;
            }
        }, 500);
    }

    DEFINE(O)
    .static("busy", busy)
    .static("idle", idle)
    .property("loading", {
        get: ()=>loading.value
    })

    return O;

};

export default BusyLoader