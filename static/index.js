// https://blog.mozilla.org/webrtc/signaling-with-rtcsimpleconnection/

let gUMConfig = { "audio": false, "video": true };
const IP_PORT_REGEX = /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+\d+/;
const RELAY_IP = "RELAY_IP_FROM_GO_SERVER";
const RELAY_PORT = "RELAY_PORT_FROM_GO_SERVER";
const IPV6_REGEX = RegExp('\:[0-9a-f]*\:[0-9a-fA-f]*','g');
const TCP_REGEX = RegExp('.*tcptype.*','g');

// Handy element access
let page = {
  get run() { return document.getElementById("run"); },
  get offer() { return document.getElementById("offer"); },
  get answer() { return document.getElementById("answer"); },
  get offerICE() { return document.getElementById("offerICE"); },
  get answerICE() { return document.getElementById("answerICE"); },
  get local() { return document.getElementById("local"); },
  get remote() { return document.getElementById("remote"); },
};


function rewrite(c, host, port) {
  c.candidate = c.candidate.replace(IP_PORT_REGEX, `${host} ${port}`);
  return c;
}

function run() {
  let offerer = new RTCPeerConnection();
  
  console.log("wtf?");

  const socket = new WebSocket('wss://localhost:' + RELAY_PORT + '/ws');

  var answer_set;
  var answer_is_set = new Promise(r => answer_set = r);

  var offer_set;
  var offer_is_set = new Promise(r => offer_set = r);

  socket.addEventListener('open', (e) => {
    offer_is_set.then((offer) => {
      console.log('Sending offer to percy');
      socket.send(offer);
    })
  })

  socket.addEventListener('message', (e) => {
    console.log('Message from percy: ', e.data);
    page.answer.value = e.data;
    answer_set(e.data);
  })

  navigator.mediaDevices.getUserMedia({video: true, audio: false})
    .then(stream => {
      console.log("got local stream");
      page.local.srcObject = stream;
      offerer.addStream(stream);
    });
  
  offerer.onicecandidate = e => {
    console.log("dropping local ICE candidate: " + JSON.stringify(e.candidate));
    return;
  }

  offerer.onnegotiationneeded = e => {
    offerer.createOffer().then(offer => {
      console.log("got local offer");
      page.offer.value = offer.sdp;
      offer_set(offer.sdp);
      return offerer.setLocalDescription(offer);
    })
    .then(() => {
      return answer_is_set;
    })
    .then((answer) => {
      console.log("setting percy's SDP answer");
      return offerer.setRemoteDescription({type: "answer", sdp: answer});
    })
    .then(() => {
      console.log("adding fake ICE candidates");
      let c = {"candidate": "candidate:0 1 UDP 2122121471 " + RELAY_IP + " " + RELAY_PORT + " typ host","sdpMid": "sdparta_0","sdpMLineIndex": 0};
      page.answerICE.value = JSON.stringify(c, null, 2) + "\n\n";
      return offerer.addIceCandidate(c);
    })
    .catch((error) => {
      console.log(error);
    })
  }
}

window.onload = () => {
  // Wire up actions
  page.run.onclick = run;

  // Clear the fields
  page.offer.value = "";
  page.answer.value = "";
  page.offerICE.value = "";
  page.answerICE.value = "";
};
