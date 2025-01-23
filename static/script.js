const roomListContainer = document.getElementById("room-list-container");
const chatContainer = document.getElementById("chat-container");
const chatInfo = document.getElementById("chat-info");
const chatBox = document.getElementById("chat-box");
const messageInput = document.getElementById("message-input");
const sendButton = document.getElementById("send-button");
const leaveButton = document.getElementById("leave-button");

let username = prompt("사용자 이름을 입력하세요:");
if (!username) username = "익명";

let selectedRoom = null;
let socket = null;

// 고정된 5개의 채팅방 목록
const rooms = [
    { roomId: "1", roomName: "GO", ownerName: "크크크크" },
    { roomId: "2", roomName: "Python", ownerName: "AI는 내가 최곤듯" },
    { roomId: "3", roomName: "Rust", ownerName: "러스트는 없나요" },
    { roomId: "4", roomName: "JAVA", ownerName: "자바가 한국에선 최고지" },
    { roomId: "5", roomName: "KOTLIN", ownerName: "코틀린 최고" }
];

// 채팅방 목록 렌더링
rooms.forEach((room) => {
    const roomElement = document.createElement("div");
    roomElement.className = "room-item";
    roomElement.setAttribute("data-room-name", room.roomName);
    roomElement.setAttribute("data-room-id", room.roomId);
    roomElement.setAttribute("data-owner-name", room.ownerName);
    roomElement.textContent = `${room.roomName} (채팅방 생성인: ${room.ownerName})`;
    roomListContainer.appendChild(roomElement);

    // 방 클릭 이벤트
    roomElement.addEventListener("click", () => {
        if (selectedRoom && selectedRoom.roomName !== room.roomName) {
            alert(`채팅방 '${selectedRoom.roomName}'에서 나갑니다.`);
        }
        selectedRoom = room;
        connectToRoom();
    });
});

// WebSocket 연결
function connectToRoom() {
    if (socket && socket.readyState === WebSocket.OPEN) {
        console.log(`기존 WebSocket 연결 종료: ${selectedRoom.roomName}`);
        socket.close();
    }

    socket = new WebSocket("ws://localhost:8080/ws");

    socket.onopen = () => {
        console.log(`채팅방 '${selectedRoom.roomName}'에 연결되었습니다.`);
        chatContainer.style.display = "flex";
        chatInfo.textContent = `채팅방: ${selectedRoom.roomName}, 채팅방 생성인: ${selectedRoom.ownerName}`;
        chatBox.innerHTML = ""; // 이전 메시지 초기화
        socket.send(JSON.stringify({
            roomId:selectedRoom.roomId,
            roomName: selectedRoom.roomName,
            ownerName: selectedRoom.ownerName,
            username: username,
            content: ""
        }));
    };

    socket.onmessage = (event) => {
        const message = JSON.parse(event.data);
        if (message.roomName === selectedRoom.roomName) {
            const messageElement = document.createElement("div");
            messageElement.className = "message";
            messageElement.innerHTML = `<strong>${message.username}:</strong> ${message.content}`;
            chatBox.appendChild(messageElement);
            chatBox.scrollTop = chatBox.scrollHeight;
        }
    };

    socket.onclose = () => {
        console.log(`WebSocket 연결 종료: ${selectedRoom.roomName}`);
        chatContainer.style.display = "none";
        selectedRoom = null;
    };
}

// 메시지 전송
sendButton.addEventListener("click", sendMessage);
messageInput.addEventListener("keypress", (event) => {
    if (event.key === "Enter") {
        event.preventDefault(); // 폼 제출 기본 동작 방지
        sendMessage();
    }
});

function sendMessage() {
    const content = messageInput.value.trim();
    if (content && socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
            roomId: selectedRoom.roomId,
            roomName: selectedRoom.roomName,
            ownerName: selectedRoom.ownerName,
            username: username,
            content: content
        }));
        messageInput.value = "";
    }
}

// 방 나가기
leaveButton.addEventListener("click", () => {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.close();
        alert(`채팅방 '${selectedRoom.roomName}'에서 나갔습니다.`);
    }
});
