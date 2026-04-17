
import {
    axios,
    URLAPIPath,
    URLHostPath,
} from "./base"


const timeline = function(camID, from, to) {
    return axios.get(URLAPIPath("cameras", camID, "timeline", from, to))
}


export default timeline;