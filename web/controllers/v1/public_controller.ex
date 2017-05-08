defmodule Condenser.API.V1.PublicController do
  use Condenser.Web, :controller
  alias Condenser.RedisWorker, as: Redis

  def meta(conn, %{"code" => code}) do
    code = String.upcase(code)
    IO.inspect code
    case Redis.get("meta/#{code}") do
      {:noexist, _} -> conn
                       |> send_resp(404, Poison.encode!(%{
                          error: "noexist",
                          message: "Code does not exist or does not have metadata."}
                        ))
      {:ok, meta}   ->
        meta_obj = Poison.decode! meta
        {:ok, url} = Redis.get(code)
        conn
        |> put_resp_content_type("application/json")
        |> send_resp(200, Poison.encode!(%{
          full_url: url,
          meta: meta_obj,
        }))
    end
  end
end