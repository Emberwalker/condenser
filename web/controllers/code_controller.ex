defmodule Condenser.CodeController do
  use Condenser.Web, :controller
  alias Condenser.RedisWorker, as: Redis

  def shortcode(conn, params) do
    case Redis.get(String.upcase(params["code"])) do
      {:noexist, _} -> conn 
                       |> put_status(:not_found)
                       |> render(Condenser.ErrorView, "404.html")
      {:ok, url}    -> conn
                       |> redirect(external: url)
    end
  end
end