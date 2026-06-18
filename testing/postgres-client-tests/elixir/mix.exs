defmodule PostgrexTest.MixProject do
  use Mix.Project

  def project do
    [
      app: :postgrex_test,
      version: "0.1.0",
      elixir: "~> 1.14",
      escript: [main_module: PostgrexTest, name: "postgrex-test"],
      deps: deps()
    ]
  end

  def application do
    [extra_applications: [:logger]]
  end

  defp deps do
    [{:postgrex, "~> 0.18"}]
  end
end
