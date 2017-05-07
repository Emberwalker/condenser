defmodule Condenser.PageController do
  use Condenser.Web, :controller

  def index(conn, _params) do
    render conn, "index.html"
  end
end
