# Content Clock Backend

This is the backend for Content Clock, a social media content scheduling and publishing tool.

## Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/content-clock-backend.git
    cd content-clock-backend
    ```

2.  **Install dependencies:**

    ```bash
    go get .
    ```

3.  **Set up environment variables:**

    Create a `.env` file in the root of the project and add the following variables. You will need to obtain the API keys and secrets from the respective social media platforms.

    ```
    # General
    API_HOST="http://localhost:8080"
    REDIRECT_HOST="http://localhost:3000"
    JWT_KEY="your-jwt-key"
    DB_MIGRATE="true"

    # Social Media
    FACEBOOK_APP_ID="your-facebook-app-id"
    FACEBOOK_SECRET="your-facebook-secret"
    TWITTER_KEY="your-twitter-key"
    TWITTER_SECRET="your-twitter-secret"
    LINKEDIN_APP_ID="your-linkedin-app-id"
    LINKEDIN_SECRET="your-linkedin-secret"
    PINTEREST_APP_ID="your-pinterest-app-id"
    PINTEREST_SECRET="your-pinterest-secret"
    MASTODON_CLIENT_KEY="your-mastodon-client-key"
    MASTODON_CLIENT_SECRET="your-mastodon-client-secret"
    MASTODON_BASE_URL="social.mastodon.social"

    # AI
    CHAT_GPT_KEY="your-chat-gpt-key"
    ```

4.  **Run the application:**

    ```bash
    go run . serve --http="localhost:8080"
    ```

    The application will be available at `http://localhost:8080`.

## Docker

You can also run the application using Docker.

1.  **Build the image:**

    ```bash
    docker build -t content-clock-backend .
    ```

2.  **Run the container:**

    ```bash
    docker run -p 8080:8080 --env-file .env content-clock-backend
    ```
