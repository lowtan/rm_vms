
import {
    axios,
    URLAPIPath,
    URLHostPath,
} from "./base"


const cameraList = function() {
    return axios.get(URLAPIPath("cameras"))
}

const CAM = {
    list: cameraList
}


export default CAM;