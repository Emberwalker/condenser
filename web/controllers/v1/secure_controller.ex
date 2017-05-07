defmodule Condenser.API.V1.SecureController do
  use Condenser.Web, :controller
  alias Condenser.RedisWorker, as: Redis
  alias Condenser.Security

  plug :check_api_key

  def shorten(conn, params) do
    raise "Not Implemented"
  end

  def delete(conn, params) do
    raise "Not Implemented"
  end

  def check_api_key(conn, _) do
    key_header = Enum.find(conn.req_headers, fn({hname, _}) -> hname == "x-api-key" end)
    case key_header do
      nil        -> conn
                    |> send_resp(401, Poison.encode!(%{
                      error: "nokey",
                      message: "No API key in X-API-Key header."}
                    ))
                    |> halt
      {_, value} ->
        sec = Security.get(value)
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
end