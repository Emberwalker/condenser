defmodule Condenser.Router do
  use Condenser.Web, :router

  pipeline :browser do
    plug :accepts, ["html"]
    plug :fetch_session
    plug :fetch_flash
    plug :protect_from_forgery
    plug :put_secure_browser_headers
    plug :put_headers, %{server: "Condenser (Phoenix/Plug)"}
  end

  pipeline :api do
    plug :accepts, ["json"]
    plug :put_headers, %{server: "Condenser (Phoenix/Plug)"}
  end

  scope "/", Condenser do
    pipe_through :browser # Use the default browser stack

    get "/", PageController, :index
  end

  # Other scopes may use custom stacks.
  scope "/api", Condenser do
    pipe_through :api
  end

  # Custom plugs
  # put_headers/2 based on http://www.phoenixframework.org/docs/understanding-plug
  def put_headers(conn, key_values) do
    Enum.reduce key_values, conn, fn {k, v}, conn ->
      Plug.Conn.put_resp_header(conn, to_string(k), v)
    end
  end
end
