const roomListContainer = document.getElementById("room-list-container");
const chatContainer = document.getElementById("chat-container");
const chatInfo = document.getElementById("chat-info");
const chatBox = document.getElementById("chat-box");
const messageInput = document.getElementById("message-input");
const sendButton = document.getElementById("send-button");
const leaveButton = document.getElementById("leave-button");
const createRoomButton = document.getElementById("create-room-button");

function loadRooms() {
    fetch("/rooms")
        .then((response) => {
            if (!response.ok) {
                throw new Error("방 목록 가져오기 실패");
            }
            return response.json();
        })
        .then((rooms) => {
            roomListContainer.innerHTML = ""; // 기존 방 목록 초기화

            rooms.forEach((room) => {
                const roomElement = document.createElement("div");
                roomElement.className = "room-item";
                roomElement.textContent = `${room.roomName}`;
                roomListContainer.appendChild(roomElement);

                // 방 클릭 이벤트 추가
                roomElement.addEventListener("click", () => {
                    selectedRoom = room;
                    connectToRoom(); // WebSocket 연결
                });
            });
        })
        .catch((error) => {
            console.error(error);
            alert("방 목록 가져오기 중 오류가 발생했습니다.");
        });
}


createRoomButton.addEventListener("click", () => {
    const roomName = prompt("생성할 방 이름을 입력하세요:");
    if (!roomName) {
        alert("방 이름을 입력해야 합니다.");
        return;
    }

    const ownerName = username; // 현재 사용자 이름을 방 생성자로 설정
    const roomId = Math.floor(Math.random() * 1000000); // 0부터 999999까지의 임의의 숫자 생성

    fetch("/create-room", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ roomId, roomName, ownerName }),
    })
        .then((response) => {
            if (!response.ok) {
                throw new Error("방 생성 실패");
            }
            return response.json();
        })
        .then(() => {
            alert(`방 '${roomName}'이 생성되었습니다.`);
            loadRooms(); // 방 목록 갱신
        })
        .catch((error) => {
            console.error(error);
            alert("방 생성 중 오류가 발생했습니다.");
        });
});

let username = prompt("사용자 이름을 입력하세요:");
if (!username) username = "익명";

let selectedRoom = null;
let socket = null;

// 고정된 5개의 채팅방 목록
const rooms = [];

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
            
        // 자신이 보낸 메시지인지 확인
        if (message.username === username) {
            // 자신의 메시지: 오른쪽 정렬, 닉네임 숨김
            messageElement.style.textAlign = "right";
            messageElement.innerHTML = `
                <span style="font-size: 0.8em; color: gray;">${message.timestamp}</span>  
                ${message.content}
            `;
        } else {
            // 다른 사용자의 메시지: 왼쪽 정렬, 닉네임 표시
            messageElement.style.textAlign = "left";
            messageElement.innerHTML = `
                <strong>${message.username}</strong> 
                <span style="font-size: 0.8em; color: gray;">${message.timestamp}</span> : 
                ${message.content}
            `;
        }

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
