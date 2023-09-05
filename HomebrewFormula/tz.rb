class Tz < Formula
  desc "Show time zones based on user defined groups"
  homepage "https://github.com/DavidGamba/dgtools/tree/master/tz"
  head "https://github.com/DavidGamba/dgtools.git", branch: "master"

  depends_on "go" => :build

  def install
    cd "tz" do
      system "go", "get"
      system "go", "build"
      bin.install "tz"
    end
    cd "HomebrewFormula" do
      inreplace "completions.bash", "tool", "tz"
      inreplace "completions.zsh", "tool", "tz"
      ohai "Installing bash completion..."
      bash_completion.install "completions.bash" => "dgtools.tz.bash"
      ohai %{Installing zsh completion...
      To enable zsh completion add this to your ~/.zshrc

      \tsource #{zsh_completion.sub prefix, HOMEBREW_PREFIX}/dgtools.tz.zsh
      }
      zsh_completion.install "completions.zsh" => "dgtools.tz.zsh"
    end
  end

  test do
    assert_match /Use 'tz help <command>' for extra details/, shell_output("#{bin}/tz")
  end
end
