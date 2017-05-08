defmodule Condenser.Random do
  @rand_symbols 'ABCDEFGHIJKLMNOPQRSTUVWXYZ23456789'

  # Based on http://stackoverflow.com/a/38315317
  @spec generate_string(integer) :: String.t
  def generate_string(len) do
    @rand_symbols
    |> Enum.take_random(len)
    |> List.to_string()
  end
end