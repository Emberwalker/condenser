defmodule Condenser.RedisWorker do
  import Logger
  use GenServer

  # Based heavily on http://elixir-lang.org/getting-started/mix-otp/supervisor-and-application.html
  def start_link(name) do
    GenServer.start_link(__MODULE__, :ok, name: name)
  end

  def init(:ok) do
    info "Redis worker starting."
    {:ok, get_client()}
  end

  @spec config_changed() :: atom
  def config_changed() do
    GenServer.call(__MODULE__, {:config_changed})
  end

  @spec exists(String.t) :: boolean
  def exists(key) do
    GenServer.call(__MODULE__, {:exists, key})
  end

  @spec set(String.t, String.t) :: atom
  def set(key, value) do
    GenServer.call(__MODULE__, {:set, key, value})
  end

  @spec get(String.t) :: {atom, String.t | nil}
  def get(key) do
    GenServer.call(__MODULE__, {:get, key})
  end
  
  def handle_call({:config_changed}, _from, client) do
    warn "Restarting Eredis client due to config change."
    Process.unlink client
    Process.exit client, :config_changed
    {:reply, :ok, get_client()}
  end

  def handle_call({:exists, key}, _from, client) do
    {:ok, res} = :eredis.q client, ['EXISTS', String.to_charlist key]
    {:reply, String.to_integer(res) == 1, client}
  end

  def handle_call({:set, key, value}, _from, client) do
    {:ok, 'OK'} = :eredis.q client, ['SET', String.to_charlist(key), String.to_charlist(value)]
    {:reply, :ok, client}
  end

  def handle_call({:get, key}, _from, client) do
    {:ok, res} = :eredis.q client, ['GET', String.to_charlist(key)]
    resp = case res do
        :undefined -> {:noexist, nil}
        x -> {:ok, x}
    end
    {:reply, resp, client}
  end

  @spec get_client() :: pid
  defp get_client() do
    [host: host, port: port] = Application.get_env(:condenser, __MODULE__)
    info "Spawning Eredis client on #{host}:#{port}."
    {:ok, pid} = :eredis.start_link(String.to_charlist(host), port)
    pid
  end
end