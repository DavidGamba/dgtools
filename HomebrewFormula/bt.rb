class Bt < Formula
  desc "A no commitments Terraform wrapper that provides build caching functionality"

  tool_name = "bt"

  homepage "https://github.com/DavidGamba/dgtools/tree/master/#{tool_name}"
  head "https://github.com/DavidGamba/dgtools.git", branch: "master"

  depends_on "go" => :build

  def install
    tool_name = "bt"

    cd "#{tool_name}" do
      system "go", "get"
      system "go", "build"
      bin.install "#{tool_name}"
    end
    cd "HomebrewFormula" do
      inreplace "completions.bash", "tool", "#{tool_name}"
      inreplace "completions.zsh", "tool", "#{tool_name}"
      ohai "Installing bash completion..."
      bash_completion.install "completions.bash" => "dgtools.#{tool_name}.bash"
      ohai %{Installing zsh completion...
      To enable zsh completion add this to your ~/.zshrc

      \tsource #{zsh_completion.sub prefix, HOMEBREW_PREFIX}/dgtools.#{tool_name}.zsh
      }
      zsh_completion.install "completions.zsh" => "dgtools.#{tool_name}.zsh"
    end
  end

  test do
    assert_match /Use '#{tool_name} help <command>' for extra details/, shell_output("#{bin}/#{tool_name} --help")
  end
end
