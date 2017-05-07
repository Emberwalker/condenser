# This file is responsible for configuring your application
# and its dependencies with the aid of the Mix.Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.
use Mix.Config

# Configures the endpoint
config :condenser, Condenser.Endpoint,
  url: [host: "localhost"],
  secret_key_base: "fcPk5zIpQm4j8Wb95Q7xH4X8NJufNghNaYwfYYVoR78592Iv/BVKZDjXRN8Y9d9b",
  render_errors: [view: Condenser.ErrorView, accepts: ~w(html json)],
  pubsub: [name: Condenser.PubSub,
           adapter: Phoenix.PubSub.PG2]

# Configures Elixir's Logger
config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Import environment specific config. This must remain at the bottom
# of this file so it overrides the configuration defined above.
import_config "#{Mix.env}.exs"

# Import site-specific config
import_config "site.exs"

