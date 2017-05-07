defmodule Condenser.Security do
  import Logger

  def start_link() do
    GenServer.start_link(__MODULE__, :ok, name: __MODULE__)
  end

  def init(:ok) do
    keys = get_config()
    info "Starting security server with #{length(keys)} keys."
    {:ok, keys}
  end

  def config_changed() do
    GenServer.call(__MODULE__, {:config_changed})
  end

  @spec get(String.t) :: {atom, {String.t, String.t} | nil}
  def get(key) do
    GenServer.call(__MODULE__, {:get, key})
  end

  def handle_call({:config_changed}, old_keys) do
    keys = get_config()
    info "Reloading security server configuration; #{length(old_keys)} keys -> #{length(keys)} keys"
    {:noreply, keys}
  end

  def handle_call({:get, key}, keys) do
    found = Enum.find(keys, fn ent -> ent.key == key end)
    res = case found do
      nil -> {:noexist, nil}
      x -> {:ok, x}
    end
    {:reply, res, keys}
  end

  defp get_config() do
    Application.get_env(:condenser, :api_keys)
  end

end
