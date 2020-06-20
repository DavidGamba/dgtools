%define _builddir          .
%define _rpmfilename       %%{NAME}-%%{VERSION}-%%{RELEASE}.%%{ARCH}.rpm
%define _source_payload    w9.lzdio
%define _binary_payload    w9.lzdio
%define _filedir           %(pwd)

Name:    cli-bookmarks
Summary: CLI Filesystem Directory Bookmarks
Version: 0.3.0
Release: 1
License: MPL-2.0
Group:   Development/Tools
Vendor:  David Gamba
Source0: cli-bookmarks
Source1: cli-bookmarks.1

%description
CLI Filesystem Directory Bookmarks

%prep
set -x
%setup -n %{buildroot} -c -T

cp %{_sourcedir}/cli-bookmarks .
cp %{_sourcedir}/cli-bookmarks.1 .

%build
set -x
if [ ! -e %{_rpmdir} ]; then
  mkdir -p %{_rpmdir}
fi

mkdir -p %{buildroot}/%{_bindir}
mkdir -p %{buildroot}/%{_mandir}/man1

cp -L %{_sourcedir}/cli-bookmarks   %{buildroot}/%{_bindir}/cli-bookmarks
cp -L %{_sourcedir}/cli-bookmarks.1 %{buildroot}/%{_mandir}/man1/cli-bookmarks.1
gzip %{buildroot}/%{_mandir}/man1/cli-bookmarks.1

rm %{buildroot}/cli-bookmarks
rm %{buildroot}/cli-bookmarks.1

%files
%attr(0755,root,root) %{_bindir}/cli-bookmarks
%attr(0644,root,root) %{_mandir}/man1/cli-bookmarks.1.gz
