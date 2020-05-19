Name: bouncer
Version: 0.0.1
Release: 1%{?dist}
Summary: Analyses exim bounce emails

License: MIT License
Source0: https://github.com/boreevyuri/bounce-collector/archive/v.%{version}/%{name}-%{version}.tar.gz
BuildArch: x86_64

%description
Analyses exim bounce emails by pipe transport and puts reusult in redis.
Also checks emails in redis and outputs "Pass" or "Decline" to use in exim router

%prep
%autosetup -n bounce-collector-v.%{version}

%install
mkdir -p %{buildroot}%{_bindir}
install -p -m 755 %{SOURCE0} %{buildroot}%{_bindir}

%files
%{_bindir}/bouncer

%changelog