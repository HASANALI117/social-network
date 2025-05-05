## Testing WebSocket Chat with Postman

Follow these steps to connect to the WebSocket endpoint, send messages, and view responses using Postman.

**Prerequisites:**

1.  **Authentication:** Before establishing a WebSocket connection, you must authenticate with the backend via a standard HTTP login request (e.g., POST to `/api/auth/signin` with credentials). Postman needs the session cookie returned by a successful login to authenticate the WebSocket connection. Ensure you have successfully logged in using Postman and that it's configured to automatically send cookies with subsequent requests.
2.  **Backend Running:** Make sure your Go backend server is running. Note the port it's running on (commonly `8080`, but check your configuration).

**Steps:**

1.  **Create a New WebSocket Request:**
    *   In Postman, click the "New" button (or `+` icon) and select "WebSocket Request".

2.  **Enter the URL:**
    *   In the URL bar, enter the WebSocket endpoint address. It uses the `ws://` scheme (or `wss://` for secure connections).
    *   The format is `ws://<your_backend_host>:<port>/ws`.
    *   Replace `<your_backend_host>` with `localhost` (or the appropriate hostname/IP if running elsewhere).
    *   Replace `<port>` with the port your backend server is listening on (e.g., `8080`).
    *   The confirmed path is `/ws`.
    *   **Example:** `ws://localhost:8080/ws`

3.  **Connect:**
    *   Click the "Connect" button. If authentication (via the session cookie) and the URL are correct, Postman will establish a connection. You should see a confirmation message in the "Messages" pane.

4.  **Send Chat Messages:**
    *   In the bottom pane (where you compose messages), select "JSON" as the format type.
    *   Enter the message payload based on the type of message you want to send:
        *   **Direct Message:**
            ```json
            {
              "type": "direct",
              "receiver_id": "USER_ID_OF_RECIPIENT",
              "content": "Hello from Postman!"
            }
            ```
            Replace `"USER_ID_OF_RECIPIENT"` with the actual ID of the user you want to message.
        *   **Group Message:**
            ```json
            {
              "type": "group",
              "receiver_id": "GROUP_ID_OF_TARGET",
              "content": "Group message from Postman!"
            }
            ```
            Replace `"GROUP_ID_OF_TARGET"` with the actual ID of the group you want to message.
    *   Click the "Send" button.

5.  **View Incoming Messages:**
    *   Messages sent *from* the server (including messages you sent, echoed back, or messages from other users) will appear in the "Messages" pane above the composer.
    *   You'll see messages you sent, messages sent by other connected clients (if any are also connected and sending to you or groups you're in), and potentially status messages like user connection/disconnection notifications. The format of received messages will likely match the `Message` struct defined in `hub.go`:
        ```json
        {
          "type": "direct/group/online_users/etc",
          "sender_id": "...",
          "receiver_id": "...",
          "content": "...",
          "created_at": "..."
        }
        ```

6.  **Disconnect:**
    *   When finished, click the "Disconnect" button.

This process allows you to simulate a WebSocket client interacting with your chat backend directly through Postman.
