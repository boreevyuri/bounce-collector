
Name: bouncer
Version: 0.0.1
Release: 1%{?dist}
Summary: Analyses exim bounce emails

License: MIT License
Source0: bouncer
Source1: bouncer.yaml
URL: https://github.com/boreevyuri/bounce-collector
BuildArch: x86_64

%description
Analyses exim bounce emails by pipe transport and puts reusult in redis.
Also checks emails in redis and outputs "Pass" or "Decline" to use in exim router

%build
echo "OK"

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}{%{_bindir},%{_sysconfdir}}
install -p -m 755 %{SOURCE0} %{buildroot}%{_bindir}
install -p -m 755 %{SOURCE1} %{buildroot}%{_sysconfdir}

%files
%defattr(0775,root,root,-)
%{_bindir}/bouncer
%{_sysconfdir}/bouncer.yaml

%changelog