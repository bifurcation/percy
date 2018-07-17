// https://blog.mozilla.org/webrtc/signaling-with-rtcsimpleconnection/

let gUMConfig = { "audio": false, "video": true };
const IP_PORT_REGEX = /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+\d+/;
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
  let pc = new RTCPeerConnection();

  console.log("wtf?");

  const socket = new WebSocket('wss://localhost:' + RELAY_PORT + '/ws');

  var answer_set;
  var answer_is_set = new Promise(r => answer_set = r);

  var offer_set;
  var offer_is_set = new Promise(r => offer_set = r);

  var ice_candidate_set;
  var ice_candidate_is_set = new Promise(r => ice_candidate_set = r);

  socket.addEventListener('open', (e) => {
    answer_is_set.then((answer) => {
      console.log('Sending SDP to percy');
      socket.send(answer);
    })
  })

  socket.addEventListener('message', (e) => {
    console.log(e.data);
    message = JSON.parse(e.data);
    if(message.type === "sdp") {
      console.log('SDP from percy: ', message.data);
      page.offer.value = message.data;
      offer_set(message.data);
    } else if(message.type === "ice") {
      console.log("ice-candidates from percy: ", message.data);
      page.answerICE.value = JSON.stringify(message.data, null, 2) + "\n\n";
      ice_candidate_set(message.data);
    }
  })

  page.offerICE.value = "ICE connection state: " + pc.iceConnectionState;

  navigator.mediaDevices.getUserMedia({video: true, audio: true})
    .then(stream => {
      console.log("got local stream");
      page.local.srcObject = stream;
      pc.addStream(stream);
    });

  pc.onicecandidate = e => {
    console.log("dropping local ICE candidate: " + JSON.stringify(e.candidate));
    return;
  }

  pc.oniceconnectionstatechange = e => {
    page.offerICE.value = page.offerICE.value + "\nICE connection state: " + pc.iceConnectionState;
  }

  pc.ontrack = e => {
    console.log("got remote track");
    page.remote.srcObject = e.streams[0];
  }

  pc.onnegotiationneeded = e => {
    offer_is_set.then((offer) => {
      console.log("setting percy's SDP");
      return pc.setRemoteDescription({type: "offer", sdp: offer});
    })
    .then(() => {
      return pc.createAnswer();
    })
    .then((answer) => {
      console.log("got local SDP");
      page.answer.value = answer.sdp;
      answer_set(answer.sdp);
      return pc.setLocalDescription(answer);
    })
    .then(() => {
      return ice_candidate_is_set;
    })
    .then((candidate) => {
      console.log("adding fake ICE candidates");
      return pc.addIceCandidate(new RTCIceCandidate(candidate));
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
