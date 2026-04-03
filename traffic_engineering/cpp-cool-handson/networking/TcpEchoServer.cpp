// Cross-platform TCP Echo Server
#include <iostream>
#include <string>
#include <cstring>
#include <thread>
#include <vector>

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

void handle_client(socket_t client_sock, const std::string& client_addr) {
    std::cout << "[Server] Client connected: " << client_addr << "\n";
    char buffer[BUFFER_SIZE];

    while (true) {
        int bytes_received = recv(client_sock, buffer, BUFFER_SIZE - 1, 0);
        if (bytes_received <= 0) {
            std::cout << "[Server] Client disconnected: " << client_addr << "\n";
            break;
        }
        buffer[bytes_received] = '\0';
        std::cout << "[Server] Received from " << client_addr << ": " << buffer;

        // Echo back
        send(client_sock, buffer, bytes_received, 0);
    }
    CLOSE_SOCKET(client_sock);
}

int main() {
#ifdef _WIN32
    WSADATA wsa_data;
    if (WSAStartup(MAKEWORD(2, 2), &wsa_data) != 0) {
        std::cerr << "WSAStartup failed\n";
        return 1;
    }
#endif

    socket_t server_sock = socket(AF_INET, SOCK_STREAM, 0);
    if (server_sock == INVALID_SOCKET) {
        std::cerr << "Failed to create socket\n";
        return 1;
    }

    // Allow address reuse
    int opt = 1;
    setsockopt(server_sock, SOL_SOCKET, SO_REUSEADDR,
               reinterpret_cast<const char*>(&opt), sizeof(opt));

    sockaddr_in server_addr{};
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = INADDR_ANY;
    server_addr.sin_port = htons(PORT);

    if (bind(server_sock, reinterpret_cast<sockaddr*>(&server_addr),
             sizeof(server_addr)) == SOCKET_ERROR) {
        std::cerr << "Bind failed\n";
        CLOSE_SOCKET(server_sock);
        return 1;
    }

    if (listen(server_sock, SOMAXCONN) == SOCKET_ERROR) {
        std::cerr << "Listen failed\n";
        CLOSE_SOCKET(server_sock);
        return 1;
    }

    std::cout << "[Server] Listening on port " << PORT << "...\n";
    std::vector<std::thread> client_threads;

    while (true) {
        sockaddr_in client_addr{};
        int addr_len = sizeof(client_addr);
        socket_t client_sock = accept(server_sock,
            reinterpret_cast<sockaddr*>(&client_addr), &addr_len);

        if (client_sock == INVALID_SOCKET) {
            std::cerr << "[Server] Accept failed\n";
            continue;
        }

        char ip_str[INET_ADDRSTRLEN];
        inet_ntop(AF_INET, &client_addr.sin_addr, ip_str, sizeof(ip_str));
        std::string addr_string = std::string(ip_str) + ":" +
                                  std::to_string(ntohs(client_addr.sin_port));

        client_threads.emplace_back(handle_client, client_sock, addr_string);
        client_threads.back().detach();
    }

    CLOSE_SOCKET(server_sock);
#ifdef _WIN32
    WSACleanup();
#endif
    return 0;
}
