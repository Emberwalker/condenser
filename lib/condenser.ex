defmodule Condenser do
  use Application

  # See http://elixir-lang.org/docs/stable/elixir/Application.html
  # for more information on OTP Applications
  def start(_type, _args) do
    import Supervisor.Spec

    # Define workers and child supervisors to be supervised
    children = [
      # Start the endpoint when the application starts
      supervisor(Condenser.Endpoint, []),
      # Start Condenser.RedisWorker via start_link/1
      worker(Condenser.RedisWorker, [Condenser.RedisWorker]),
    ]

    # See http://elixir-lang.org/docs/stable/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: Condenser.Supervisor]
    Supervisor.start_link(children, opts)
  end

  # Tell Phoenix to update the endpoint configuration
  # whenever the application is updated.
  def config_change(changed, _new, removed) do
    Condenser.Endpoint.config_change(changed, removed)
    Condenser.RedisWorker.config_changed()
    :ok
  end
end