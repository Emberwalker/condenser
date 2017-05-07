defmodule Condenser.Router do
  use Condenser.Web, :router

  pipeline :browser do
    plug :accepts, ["html"]
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
    get "/:code", CodeController, :shortcode
  end

  # Other scopes may use custom stacks.
  scope "/api", Condenser.API do
    pipe_through :api

    get "/meta/:code", V1.PublicController, :meta
    post "/shorten", V1.SecureController, :shorten
    post "/delete", V1.SecureController, :delete

    scope "/v1", V1, as: :v1 do
      get "/meta/:code", PublicController, :meta
      post "/shorten", SecureController, :shorten
      post "/delete", SecureController, :delete
    end
  end

  # Custom plugs
  # put_headers/2 based on http://www.phoenixframework.org/docs/understanding-plug
  def put_headers(conn, key_values) do
    Enum.reduce key_values, conn, fn {k, v}, conn ->
      Plug.Conn.put_resp_header(conn, to_string(k), v)
    end
  end
end
