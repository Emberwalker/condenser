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

  def check_api_key(conn, _) do
    import Plug.Conn
    key_header = Enum.find(conn.req_headers, fn({hname, _}) -> hname == "x-api-key" end)
    case key_header do
      nil        -> conn
                    |> send_resp(401, Poison.encode!(%{
                      error: "nokey",
                      message: "No API key in X-API-Key header."}
                    ))
                    |> halt
      {_, value} ->
        sec = get(value)
        case sec do
          {:noexist, _}   -> conn
                             |> send_resp(401, Poison.encode!(%{
                               error: "invalidkey",
                               message: "Invalid API key in X-API-Key header."}
                             ))
                             |> halt
          {:ok, sec_info} -> assign(conn, :user, sec_info.name)
        end
    end
  end

  def handle_call({:config_changed}, _from, old_keys) do
    keys = get_config()
    info "Reloading security server configuration; #{length(old_keys)} keys -> #{length(keys)} keys"
    {:noreply, keys}
  end

  def handle_call({:get, user_key}, _from, keys) do
    found = Enum.find(keys, fn(ex) -> ex.key == user_key end)
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
