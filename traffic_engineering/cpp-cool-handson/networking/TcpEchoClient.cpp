// Cross-platform TCP Echo Client
#include <iostream>
#include <string>
#include <cstring>

#ifdef _WIN32
    #include <winsock2.h>
    #include <ws2tcpip.h>
    #pragma comment(lib, "ws2_32.lib")
    using socket_t = SOCKET;
    #define CLOSE_SOCKET closesocket
#else
    #include <sys/socket.h>
    #include <netinet/in.h>
    #include <arpa/inet.h>
    #include <unistd.h>
    using socket_t = int;
    #define INVALID_SOCKET -1
    #define SOCKET_ERROR -1
    #define CLOSE_SOCKET close
#endif

constexpr int PORT = 9090;
constexpr int BUFFER_SIZE = 1024;

int main() {
#ifdef _WIN32
    WSADATA wsa_data;
    if (WSAStartup(MAKEWORD(2, 2), &wsa_data) != 0) {
        std::cerr << "WSAStartup failed\n";
        return 1;
    }
#endif

    socket_t sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock == INVALID_SOCKET) {
        std::cerr << "Failed to create socket\n";
        return 1;
    }

    sockaddr_in server_addr{};
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(PORT);
    inet_pton(AF_INET, "127.0.0.1", &server_addr.sin_addr);

    if (connect(sock, reinterpret_cast<sockaddr*>(&server_addr),
                sizeof(server_addr)) == SOCKET_ERROR) {
        std::cerr << "Connection failed\n";
        CLOSE_SOCKET(sock);
        return 1;
    }

    std::cout << "[Client] Connected to server on port " << PORT << "\n";
    std::cout << "[Client] Type messages (Ctrl+C to quit):\n";

    std::string line;
    char buffer[BUFFER_SIZE];

    while (std::getline(std::cin, line)) {
        line += "\n";
        send(sock, line.c_str(), static_cast<int>(line.size()), 0);

        int bytes = recv(sock, buffer, BUFFER_SIZE - 1, 0);
        if (bytes <= 0) {
            std::cout << "[Client] Server disconnected\n";
            break;
        }
        buffer[bytes] = '\0';
        std::cout << "[Client] Echo: " << buffer;
    }

    CLOSE_SOCKET(sock);
#ifdef _WIN32
    WSACleanup();
#endif
    return 0;
}
