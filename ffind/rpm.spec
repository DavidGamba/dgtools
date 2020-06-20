%define _builddir          .
%define _rpmfilename       %%{NAME}-%%{VERSION}-%%{RELEASE}.%%{ARCH}.rpm
%define _source_payload    w9.lzdio
%define _binary_payload    w9.lzdio
%define _filedir           %(pwd)

Name:    ffind
Summary: Fast Regex Find, faster common Find searches using the power of Golang’s Regex engine
Version: 0.5.0
Release: 1
License: MPL-2.0
Group:   Development/Tools
Vendor:  David Gamba
Source0: ffind
Source1: ffind.1

%description
Fast Regex Find, faster common Find searches using the power of Golang’s Regex engine

%prep
set -x
%setup -n %{buildroot} -c -T

cp %{_sourcedir}/ffind .
cp %{_sourcedir}/ffind.1 .

%build
set -x
if [ ! -e %{_rpmdir} ]; then
  mkdir -p %{_rpmdir}
fi

mkdir -p %{buildroot}/%{_bindir}
mkdir -p %{buildroot}/%{_mandir}/man1

cp -L %{_sourcedir}/ffind   %{buildroot}/%{_bindir}/ffind
cp -L %{_sourcedir}/ffind.1 %{buildroot}/%{_mandir}/man1/ffind.1
gzip %{buildroot}/%{_mandir}/man1/ffind.1

rm %{buildroot}/ffind
rm %{buildroot}/ffind.1

%files
%attr(0755,root,root) %{_bindir}/ffind
%attr(0644,root,root) %{_mandir}/man1/ffind.1.gz
