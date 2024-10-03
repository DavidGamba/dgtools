class Yamlseam < Formula
  @@tool_name = "yaml-seam"
  @@tool_desc = "Parses YAML input passed from file(s) or piped to STDIN and allows to split it or combine it"
  @@tool_path = "yaml-seam"

  desc "#{@@tool_desc}"
  homepage "https://github.com/DavidGamba/dgtools/tree/master/#{@@tool_name}"
  url "https://github.com/DavidGamba/dgtools/archive/refs/tags/yaml-seam/v0.1.0.tar.gz"
  sha256 "0818d14a3a0cfbac8077f973eae44e00315a8440ab924a9060928367bca59b7f"

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
