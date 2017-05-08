defmodule Condenser.API.V1.SecureController do
  use Condenser.Web, :controller
  import Condenser.Security, only: [check_api_key: 2]
  import Condenser.Random, only: [generate_string: 1]
  alias Condenser.RedisWorker, as: Redis

  plug :check_api_key

  def shorten(conn, %{"url" => url, "code" => code}) do
    do_shorten(conn, url, code) |> shorten_response(conn)
  end

  def shorten(conn, %{"url" => url}) do
    shorten_until_success(conn, url) |> shorten_response(conn)
  end

  defp shorten_response(reply, conn) do
    case reply do
      {:conflict, _} -> conn |> send_resp(409, Poison.encode!(%{
        "error" => "conflict",
        "message" => "Code already exists."
      }))
      {:ok, code} -> conn |> send_resp(200, Poison.encode!(%{
        "short_url" => get_url(conn, code)
      }))
    end
  end

  @spec shorten_until_success(Plug.Conn.t, String.t, integer) :: {atom, String.t | nil}
  defp shorten_until_success(conn, url, attempts \\ 0) do
    if attempts > 10 do
      {:conflict, nil}
    else
      code = generate_string(Application.get_env(:condenser, :key_length, 6))
      case do_shorten(conn, url, code) do
        {:conflict, _} -> shorten_until_success(conn, url, attempts + 1)
        {:ok, new_code} -> {:ok, new_code}
      end
    end
  end

  @spec do_shorten(Plug.Conn.t, String.t, String.t) :: {atom, String.t | nil}
  defp do_shorten(conn, url, code) do
    code = String.upcase(code)
    if Redis.exists(code) do
      {:conflict, nil}
    else
      meta = %{
        "owner" => conn.assigns.user,
        "time" => DateTime.utc_now()
      }
      meta = case conn.params["meta"] do
        nil -> meta
        x -> Map.put(meta, "user_meta", x)
      end
      meta = Poison.encode! meta
      result = Redis.set(code, url)
      _meta_result = Redis.set("meta/#{code}", meta)
      {result, code}
    end
  end

  @spec get_url(Plug.Conn.t, String.t) :: String.t
  defp get_url(conn, code) do
    code = String.upcase(code)
    scheme = conn.scheme
    port = case conn.port do
      80 when scheme == :http -> ""
      443 when scheme == :https -> ""
      _ -> ":#{conn.port}"
    end
    "#{to_string(scheme)}://#{conn.host}#{port}/#{code}"
  end

  def delete(conn, %{"code" => code}) do
    code = String.upcase(code)
    reply = if Redis.exists(code) do
      code_del = Redis.del(code)
      _meta_del = Redis.del("meta/#{code}")
      to_string(code_del)
    else
      "noexist"
    end
    conn |> send_resp(200, Poison.encode!(%{
      "code" => code,
      "status" => reply
    }))
  end

end