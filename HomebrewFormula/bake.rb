class Bake < Formula
  @@tool_name = "bake"
  @@tool_desc = "Go Build + Something like Make = Bake ¯\\_(ツ)_/¯"
  @@tool_path = "bake"

  desc "#{@@tool_desc}"
  homepage "https://github.com/DavidGamba/go-getoptions/tree/master/#{@@tool_name}"
  url "https://github.com/DavidGamba/dgtools/archive/refs/tags/bake/v0.1.0.tar.gz"
  sha256 "c03ea914b3dfb885bd4821118b89411a06b7ac17f9e78b4cf066a4826a85841f"

  depends_on "go" => :build

  def install
    cd "#{@@tool_path}" do
      ENV["GOEXPERIMENT"] = "rangefunc"
      system "go", "get"
      system "go", "build"
      bin.install "#{@@tool_name}"
    end
    cd "HomebrewFormula" do
      inreplace "completions.bash", "tool", "#{@@tool_name}"
      inreplace "completions.zsh", "tool", "#{@@tool_name}"
      ohai "Installing bash completion..."
      bash_completion.install "completions.bash" => "dgtools.#{@@tool_name}.bash"
      ohai %{Installing zsh completion...
      To enable zsh completion add this to your ~/.zshrc

      \tsource #{zsh_completion.sub prefix, HOMEBREW_PREFIX}/dgtools.#{@@tool_name}.zsh
      }
      zsh_completion.install "completions.zsh" => "dgtools.#{@@tool_name}.zsh"
      ohai "Installed #{@@tool_name} from #{@@tool_path} dir"
    end
  end

  test do
    assert_match /Use '#{@@tool_name} help[^']*' for extra details/, shell_output("#{bin}/#{@@tool_name} --help")
  end
end
