
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
mkdir -p %{buildroot}{%{_bindir},%{_sysconfdir}}
install -p -m 755 %{SOURCE0} %{buildroot}%{_bindir}
install -p -m 644 %{SOURCE1} %{buildroot}%{_sysconfdir}

%files
%defattr(0775,root,root,-)
%{_bindir}/bouncer
%attr(0644, root, root) %config(noreplace) %{_sysconfdir}/bouncer.yaml

%changelog
* Fri May 22 2020 Boreev Yuri <boreevyuri@gmail.com>
- deep alpha