FROM rabbitmq:3.13-management

# Install dependencies
RUN apt-get update && apt-get install -y wget

# Download the delayed message exchange plugin
ARG PLUGIN_VERSION=3.13.0
RUN wget https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v${PLUGIN_VERSION}/rabbitmq_delayed_message_exchange-${PLUGIN_VERSION}.ez \
    -O /opt/rabbitmq/plugins/rabbitmq_delayed_message_exchange-${PLUGIN_VERSION}.ez

# Enable the plugin
RUN rabbitmq-plugins enable rabbitmq_delayed_message_exchange
