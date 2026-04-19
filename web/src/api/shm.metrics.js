
import {
    axios,
    URLHostPath,
} from "./base"


const shmMetrics = function() {
    return axios.get(URLHostPath("health", "shm", "metrics"))
}


export default shmMetrics;