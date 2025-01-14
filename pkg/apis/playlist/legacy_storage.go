package playlist

import (
	"context"
	"errors"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/grafana/grafana/pkg/services/grafana-apiserver/endpoints/request"
	"github.com/grafana/grafana/pkg/services/playlist"
)

var (
	_ rest.Scoper               = (*legacyStorage)(nil)
	_ rest.SingularNameProvider = (*legacyStorage)(nil)
	_ rest.Getter               = (*legacyStorage)(nil)
	_ rest.Lister               = (*legacyStorage)(nil)
	_ rest.Storage              = (*legacyStorage)(nil)
	_ rest.Creater              = (*legacyStorage)(nil)
	_ rest.Updater              = (*legacyStorage)(nil)
	_ rest.GracefulDeleter      = (*legacyStorage)(nil)
)

type legacyStorage struct {
	service        playlist.Service
	namespacer     request.NamespaceMapper
	tableConverter rest.TableConvertor

	DefaultQualifiedResource  schema.GroupResource
	SingularQualifiedResource schema.GroupResource
}

func (s *legacyStorage) New() runtime.Object {
	return &Playlist{}
}

func (s *legacyStorage) Destroy() {}

func (s *legacyStorage) NamespaceScoped() bool {
	return true // namespace == org
}

func (s *legacyStorage) GetSingularName() string {
	return "playlist"
}

func (s *legacyStorage) NewList() runtime.Object {
	return &PlaylistList{}
}

func (s *legacyStorage) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return s.tableConverter.ConvertToTable(ctx, object, tableOptions)
}

func (s *legacyStorage) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	// TODO: handle fetching all available orgs when no namespace is specified
	// To test: kubectl get playlists --all-namespaces
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, err
	}

	limit := 100
	if options.Limit > 0 {
		limit = int(options.Limit)
	}
	res, err := s.service.Search(ctx, &playlist.GetPlaylistsQuery{
		OrgId: info.OrgID,
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

	list := &PlaylistList{}
	for _, v := range res {
		p, err := s.service.Get(ctx, &playlist.GetPlaylistByUidQuery{
			UID:   v.UID,
			OrgId: info.OrgID,
		})
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, *convertToK8sResource(p, s.namespacer))
	}
	if len(list.Items) == limit {
		list.Continue = "<more>" // TODO?
	}
	return list, nil
}

func (s *legacyStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, err
	}

	dto, err := s.service.Get(ctx, &playlist.GetPlaylistByUidQuery{
		UID:   name,
		OrgId: info.OrgID,
	})
	if err != nil || dto == nil {
		if errors.Is(err, playlist.ErrPlaylistNotFound) || err == nil {
			err = k8serrors.NewNotFound(s.SingularQualifiedResource, name)
		}
		return nil, err
	}

	return convertToK8sResource(dto, s.namespacer), nil
}

func (s *legacyStorage) Create(ctx context.Context,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, err
	}

	p, ok := obj.(*Playlist)
	if !ok {
		return nil, fmt.Errorf("expected playlist?")
	}
	cmd, err := convertToLegacyUpdateCommand(p, info.OrgID)
	if err != nil {
		return nil, err
	}
	out, err := s.service.Create(ctx, &playlist.CreatePlaylistCommand{
		UID:      p.Name,
		Name:     cmd.Name,
		Interval: cmd.Interval,
		Items:    cmd.Items,
		OrgId:    cmd.OrgId,
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, out.UID, nil)
}

func (s *legacyStorage) Update(ctx context.Context,
	name string,
	objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc,
	updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, false, err
	}

	created := false
	old, err := s.Get(ctx, name, nil)
	if err != nil {
		return old, created, err
	}

	obj, err := objInfo.UpdatedObject(ctx, old)
	if err != nil {
		return old, created, err
	}
	p, ok := obj.(*Playlist)
	if !ok {
		return nil, created, fmt.Errorf("expected playlist after update")
	}

	cmd, err := convertToLegacyUpdateCommand(p, info.OrgID)
	if err != nil {
		return old, created, err
	}
	_, err = s.service.Update(ctx, cmd)
	if err != nil {
		return nil, false, err
	}

	r, err := s.Get(ctx, name, nil)
	return r, created, err
}

// GracefulDeleter
func (s *legacyStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	v, err := s.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return v, false, err // includes the not-found error
	}
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, false, err
	}
	p, ok := v.(*Playlist)
	if !ok {
		return v, false, fmt.Errorf("expected a playlist response from Get")
	}
	err = s.service.Delete(ctx, &playlist.DeletePlaylistCommand{
		UID:   name,
		OrgId: info.OrgID,
	})
	return p, true, err // true is instant delete
}
