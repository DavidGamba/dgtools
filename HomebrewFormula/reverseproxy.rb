class Reverseproxy < Formula
  @@tool_name = "reverseproxy"
  @@tool_desc = "reverse proxy multiple services with their own base paths"
  @@tool_path = "reverseproxy"

  desc "#{@@tool_desc}"
  homepage "https://github.com/DavidGamba/dgtools/tree/master/#{@@tool_name}"
  url "https://github.com/DavidGamba/dgtools/archive/refs/tags/reverseproxy/v0.1.0.tar.gz"
  sha256 "858c1ad49859ad8672fd51eef74d8ff667bd9d699c905a6071f5fca296e1df0b"

  depends_on "go" => :build

  def install
    cd "#{@@tool_path}" do
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
