class Kdecode < Formula
  @@tool_name = "kdecode"
  @@tool_desc = "Decodes K8s secret's data block"
  @@tool_path = "kdecode"

  desc "#{@@tool_desc}"
  homepage "https://github.com/DavidGamba/dgtools/tree/master/#{@@tool_name}"
  url "https://github.com/DavidGamba/dgtools/archive/refs/tags/kdecode/v0.2.0.tar.gz"
  sha256 "12f8f4e9f0d61522ed75878cda2c6a9ef54bf298cf7e095862303c3c1d81a7b3"

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
