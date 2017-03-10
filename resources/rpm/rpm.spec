# SPEC file

%global c_vendor    %{_vendor}
%global gh_owner    %{_owner}
%global gh_project  %{_project}

Name:      %{_package}
Version:   %{_version}
Release:   %{_release}%{?dist}
Summary:   NATS Test Component.

Group:     Applications/Services
License:   Apache-2.0
URL:       https://github.com/%{gh_owner}/%{gh_project}

BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-%(%{__id_u} -n)

Provides:  %{gh_project} = %{version}

Requires(preun): chkconfig
Requires(preun): initscripts
Requires(postun): initscripts

%description
This program provides a RESTful HTTP JSON API service to test services connected to a NATS bus.
It has been designed to test any system composed by services that exchange JSON messages via a NATS bus.

%build
#(cd %{_current_directory} && make build)

%install
rm -rf $RPM_BUILD_ROOT
(cd %{_current_directory} && make install DESTDIR=$RPM_BUILD_ROOT)

%clean
rm -rf $RPM_BUILD_ROOT
(cd %{_current_directory} && make clean)

%files
%attr(-,root,root) %{_binpath}/%{_project}
%attr(-,root,root) %{_binpath}/md5str.sh
%attr(-,root,root) %{_initpath}/%{_project}
%attr(-,root,root) %{_docpath}
%attr(-,root,root) %{_manpath}/%{_project}.1.gz
%docdir %{_docpath}
%docdir %{_manpath}
%config(noreplace) %{_configpath}*

%preun
if [ $1 -eq 0 ] ; then
    # uninstall: stop service
    /sbin/service %{gh_project} stop >/dev/null 2>&1
    /sbin/chkconfig --del %{gh_project}
fi

%postun
if [ $1 -eq 1 ] ; then
    # upgrade: restart service if was running
    /sbin/service %{gh_project} condrestart >/dev/null 2>&1 || :
fi

%changelog
* Wed Jul 06 2016 Nicola Asuni <nicola.asuni@miracl.com> 5.1.5-1
- Updated license - first public version
* Sat May 28 2016 Nicola Asuni <nicola.asuni@miracl.com> 2.2.0-1
- Updated description
- Add init scripts
* Fri Apr 15 2016 Nicola Asuni <nicola.asuni@miracl.com> 1.0.0-1
- Initial Commit

